package repositories

import (
	"booking-service/config"
	"booking-service/internal/module/booking/models/entity"
	"booking-service/internal/module/booking/models/response"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	circuit "github.com/rubyist/circuitbreaker"
)

type repositories struct {
	db             *sqlx.DB
	log            log.Logger
	httpClient     *circuit.HTTPClient
	cfgUserService *config.UserServiceConfig
	redisClient    *redis.Client
}

type Repositories interface {
	// http
	ValidateToken(ctx context.Context, token string) (response.UserServiceValidate, error)
	// redis
	CheckStockTicket(ctx context.Context, ticketDetailID int64) (int64, error)
	DecrementStockTicket(ctx context.Context, ticketDetailID int64) error
	// db
	UpsertBooking(ctx context.Context, booking entity.Booking) error
	UpsertPayment(ctx context.Context, payment entity.Payment) error
	FindBookingByUserID(ctx context.Context, userID int64) (entity.Booking, error)
	FindPaymentByBookingID(ctx context.Context, bookingID int64) (entity.Payment, error)
}

func New(db *sqlx.DB, log log.Logger, httpClient *circuit.HTTPClient, redisClient *redis.Client) Repositories {
	return &repositories{
		db:          db,
		log:         log,
		httpClient:  httpClient,
		redisClient: redisClient,
	}
}

// CheckStockTicket implements Repositories.
func (r *repositories) CheckStockTicket(ctx context.Context, ticketDetailID int64) (int64, error) {
	ticketIDString := fmt.Sprintf("%d", ticketDetailID)
	data, err := r.redisClient.Get(ctx, ticketIDString).Result()
	if err != nil {
		return 0, errors.InternalServerError("error get stock ticket")
	}
	dataInt, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, errors.InternalServerError("error parse stock ticket")
	}
	return dataInt, nil
}

// DecrementStockTicket implements Repositories.
func (r *repositories) DecrementStockTicket(ctx context.Context, ticketDetailID int64) error {
	ticketIDString := fmt.Sprintf("%", ticketDetailID)
	_, err := r.redisClient.Decr(ctx, ticketIDString).Result()
	if err != nil {
		return errors.InternalServerError("error decrement stock ticket")
	}
	return nil
}

// UpsertBooking implements Repositories.
func (r *repositories) UpsertBooking(ctx context.Context, booking entity.Booking) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.InternalServerError("error starting transaction")
	}

	// Lock the rows for update
	query := `SELECT * FROM booking WHERE id = ? FOR UPDATE`
	var existingBooking entity.Booking
	err = r.db.GetContext(ctx, &existingBooking, query, booking.ID)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return errors.InternalServerError("error locking rows")
	}

	// Perform the upsert operation
	if err == sql.ErrNoRows {
		// Insert new booking
		_, err = tx.NamedExecContext(ctx, `
			INSERT INTO booking (id, user_id, ...) 
			VALUES (:id, :user_id, ...)
		`, booking)
	} else {
		// Update existing booking
		_, err = tx.NamedExecContext(ctx, `
			UPDATE booking 
			SET user_id = :user_id, ...
			WHERE id = :id
		`, booking)
	}
	if err != nil {
		tx.Rollback()
		return errors.InternalServerError("error upserting booking")
	}

	err = tx.Commit()
	if err != nil {
		return errors.InternalServerError("error committing transaction")
	}

	return nil
}

// UpsertPayment implements Repositories.
func (r *repositories) UpsertPayment(ctx context.Context, payment entity.Payment) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.InternalServerError("error starting transaction")
	}

	// Lock the rows for update
	query := `SELECT * FROM payment WHERE booking_id = ? FOR UPDATE`
	var existingPayment entity.Payment
	err = r.db.GetContext(ctx, &existingPayment, query, payment.BookingID)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return errors.InternalServerError("error locking rows")
	}

	// Perform the upsert operation
	if err == sql.ErrNoRows {
		// Insert new payment
		_, err = tx.NamedExecContext(ctx, `
			INSERT INTO payment (booking_id, ...) 
			VALUES (:booking_id, ...)
		`, payment)
	} else {
		// Update existing payment
		_, err = tx.NamedExecContext(ctx, `
			UPDATE payment 
			SET ...
			WHERE booking_id = :booking_id
		`, payment)
	}
	if err != nil {
		tx.Rollback()
		return errors.InternalServerError("error upserting payment")
	}

	err = tx.Commit()
	if err != nil {
		return errors.InternalServerError("error committing transaction")
	}

	return nil
}

// FindBookingByUserID implements Repositories.
func (r *repositories) FindBookingByUserID(ctx context.Context, userID int64) (entity.Booking, error) {
	query := `SELECT * FROM booking WHERE user_id = ?`
	var booking entity.Booking
	err := r.db.Get(&booking, query, userID)
	if err == sql.ErrNoRows {
		return entity.Booking{}, nil
	}
	if err != nil {
		return entity.Booking{}, errors.InternalServerError("error find booking by user id")
	}
	return booking, nil
}

// FindPaymentByBookingID implements Repositories.
func (r *repositories) FindPaymentByBookingID(ctx context.Context, bookingID int64) (entity.Payment, error) {
	query := `SELECT * FROM payment WHERE booking_id = ?`
	var payment entity.Payment
	err := r.db.Get(&payment, query, bookingID)
	if err == sql.ErrNoRows {
		return entity.Payment{}, nil
	}
	if err != nil {
		return entity.Payment{}, errors.InternalServerError("error find payment by booking id")
	}
	return payment, nil
}

func (r *repositories) ValidateToken(ctx context.Context, token string) (response.UserServiceValidate, error) {
	// http call to user service
	url := fmt.Sprintf("http://%s:%s/api/private/token/validate?token=%s", r.cfgUserService.Host, r.cfgUserService.Port, token)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return response.UserServiceValidate{
			IsValid: false,
			UserID:  0,
		}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		r.log.Error(ctx, "Invalid token", resp.StatusCode)
		return response.UserServiceValidate{
			IsValid: false,
			UserID:  0,
		}, errors.BadRequest("Invalid token")
	}

	// parse response
	var respData response.UserServiceValidate

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respData); err != nil {
		return response.UserServiceValidate{
			IsValid: false,
			UserID:  0,
		}, err
	}

	if !respData.IsValid {
		r.log.Error(ctx, "Invalid token", resp.StatusCode)
		return response.UserServiceValidate{
			IsValid: false,
			UserID:  0,
		}, errors.BadRequest("Invalid token")
	}

	// validate token
	return response.UserServiceValidate{
		IsValid: respData.IsValid,
		UserID:  respData.UserID,
	}, nil
}
