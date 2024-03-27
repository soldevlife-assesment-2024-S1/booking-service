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

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	circuit "github.com/rubyist/circuitbreaker"
)

type repositories struct {
	db               *sqlx.DB
	log              log.Logger
	httpClient       *circuit.HTTPClient
	cfgTicketService *config.TicketServiceConfig
	cfgUserService   *config.UserServiceConfig
	redisClient      *redis.Client
}

// FindBookingByID implements Repositories.
func (r *repositories) FindBookingByID(ctx context.Context, bookingID string) (entity.Booking, error) {
	query := `SELECT * FROM booking WHERE id = ?`
	var booking entity.Booking
	err := r.db.Get(&booking, query, bookingID)
	if err == sql.ErrNoRows {
		return entity.Booking{}, nil
	}
	if err != nil {
		return entity.Booking{}, errors.InternalServerError("error find booking by booking id")
	}
	return booking, nil
}

// InquiryTicketAmount implements Repositories.
func (r *repositories) InquiryTicketAmount(ctx context.Context, ticketDetailID int64, totalTicket int) (float64, error) {
	url := fmt.Sprintf("http://%s:%s/api/private/ticket/inquiry?ticket_detail_id=%d&total_ticket=%d", r.cfgTicketService.Host, r.cfgTicketService.Port, ticketDetailID, totalTicket)
	fmt.Println(url)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return 0, errors.InternalServerError("error inquiry ticket amount")
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		r.log.Error(ctx, "Inquiry ticket amount failed", resp.StatusCode)
		return 0, errors.BadRequest("Inquiry ticket amount failed")
	}

	// parse response
	var respBase response.BaseResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respBase); err != nil {
		return 0, err
	}

	respBase.Data = respBase.Data.(map[string]interface{})
	respData := response.InquiryTicketAmount{
		TotalTicket: int(respBase.Data.(map[string]interface{})["total_ticket"].(float64)),
		TotalAmount: respBase.Data.(map[string]interface{})["total_amount"].(float64),
	}

	return respData.TotalAmount, nil
}

type Repositories interface {
	// http
	ValidateToken(ctx context.Context, token string) (response.UserServiceValidate, error)
	InquiryTicketAmount(ctx context.Context, ticketDetailID int64, totalTicket int) (float64, error)
	// redis
	CheckStockTicket(ctx context.Context, ticketDetailID int64) (int64, error)
	DecrementStockTicket(ctx context.Context, ticketDetailID int64) error
	// db
	UpsertBooking(ctx context.Context, booking *entity.Booking) (id string, err error)
	UpsertPayment(ctx context.Context, payment *entity.Payment) error
	FindBookingByUserID(ctx context.Context, userID int64) (entity.Booking, error)
	FindBookingByID(ctx context.Context, bookingID string) (entity.Booking, error)
	FindPaymentByBookingID(ctx context.Context, bookingID string) (entity.Payment, error)
}

func New(db *sqlx.DB, log log.Logger, httpClient *circuit.HTTPClient, redisClient *redis.Client, cfgUserService *config.UserServiceConfig, cfgTicketService *config.TicketServiceConfig) Repositories {
	return &repositories{
		db:               db,
		log:              log,
		httpClient:       httpClient,
		redisClient:      redisClient,
		cfgUserService:   cfgUserService,
		cfgTicketService: cfgTicketService,
	}
}

// FindBookingByBookingID implements Repositories.
func (r *repositories) FindBookingByBookingID(ctx context.Context, bookingID string) (entity.Booking, error) {
	query := fmt.Sprintf(`SELECT * FROM bookings WHERE id = %s`, bookingID)
	var booking entity.Booking
	err := r.db.Get(&booking, query)
	if err == sql.ErrNoRows {
		return entity.Booking{}, nil
	}
	if err != nil {
		return entity.Booking{}, errors.InternalServerError("error find booking by booking id")
	}
	return booking, nil
}

// CheckStockTicket implements Repositories.
func (r *repositories) CheckStockTicket(ctx context.Context, ticketDetailID int64) (int64, error) {
	ticketIDString := fmt.Sprintf("%d", ticketDetailID)
	data, err := r.redisClient.Get(ctx, ticketIDString).Result()
	if err != nil {
		// hit api ticket service to get stock ticket
		url := fmt.Sprintf("http://%s:%s/api/private/ticket/stock?ticket_detail_id=%d", r.cfgTicketService.Host, r.cfgTicketService.Port, ticketDetailID)
		resp, err := r.httpClient.Get(url)
		if err != nil {
			return 0, errors.InternalServerError("error get stock ticket")
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			r.log.Error(ctx, "Get stock ticket failed", resp.StatusCode)
			return 0, errors.BadRequest("Get stock ticket failed")
		}

		// parse response
		var respBase response.BaseResponse

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&respBase); err != nil {
			return 0, err
		}

		respBase.Data = respBase.Data.(map[string]interface{})
		data = fmt.Sprintf("%d", int64(respBase.Data.(map[string]interface{})["stock"].(float64)))

		// set stock ticket to redis

		_, err = r.redisClient.Set(ctx, ticketIDString, data, 0).Result()
		if err != nil {
			return 0, errors.InternalServerError("error set stock ticket")
		}
	}
	dataInt, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, errors.InternalServerError("error parse stock ticket")
	}
	return dataInt, nil
}

// DecrementStockTicket implements Repositories.
func (r *repositories) DecrementStockTicket(ctx context.Context, ticketDetailID int64) error {
	ctx = context.Background()
	ticketIDString := fmt.Sprintf("%d", ticketDetailID)
	_, err := r.redisClient.Decr(ctx, ticketIDString).Result()
	if err != nil {
		return errors.InternalServerError("error decrement stock ticket")
	}
	return nil
}

// UpsertBooking implements Repositories.
func (r *repositories) UpsertBooking(ctx context.Context, booking *entity.Booking) (string, error) {
	ctx = context.Background()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", errors.InternalServerError("error starting transaction")
	}

	// Lock the rows for update
	query := `SELECT * FROM bookings WHERE id = $1 FOR UPDATE`
	var existingBooking entity.Booking
	err = r.db.GetContext(ctx, &existingBooking, query, booking.ID)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return "", errors.InternalServerError("error locking rows")
	}

	var ID string

	// Perform the upsert operation
	if err == sql.ErrNoRows {
		// Insert new booking
		queryInsert := fmt.Sprintf(`
			INSERT INTO bookings (id, user_id, ticket_detail_id, total_tickets, full_name, personal_id, booking_date) 
			VALUES ('%s', %d, %d, %d, '%s', '%s', '%s') RETURNING id
		`, booking.ID, booking.UserID, booking.TicketDetailID, booking.TotalTickets, booking.FullName, booking.PersonalID, booking.BookingDate.Format("2006-01-02 15:04:05"))

		err := tx.QueryRowContext(ctx,
			queryInsert).Scan(&ID)
		if err != nil {
			tx.Rollback()
			return "", errors.InternalServerError("error upserting booking")
		}
	} else {
		// Update existing booking
		queryUpdate := fmt.Sprintf(`
			UPDATE bookings
			SET user_id = %d, ticket_detail_id = %d, total_tickets = %d, full_name = '%s', personal_id = '%s', booking_date = '%s'
			WHERE id = '%s' RETURNING id
		`, booking.UserID, booking.TicketDetailID, booking.TotalTickets, booking.FullName, booking.PersonalID, booking.BookingDate.Format("2006-01-02 15:04:05"), booking.ID)
		err := tx.QueryRowContext(ctx, queryUpdate).Scan(&ID)
		if err != nil {
			tx.Rollback()
			return "", errors.InternalServerError("error upserting booking")
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", errors.InternalServerError("error committing transaction")
	}

	return ID, nil
}

// UpsertPayment implements Repositories.
func (r *repositories) UpsertPayment(ctx context.Context, payment *entity.Payment) error {
	ctx = context.Background()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		fmt.Println("err msg 1", err)
		return errors.InternalServerError("error starting transaction")
	}

	// Lock the rows for update
	query := `SELECT * FROM payments WHERE booking_id = $1 FOR UPDATE`
	var existingPayment entity.Payment
	err = r.db.GetContext(ctx, &existingPayment, query, payment.BookingID)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("err msg 1.5", err)
		tx.Rollback()
		return errors.InternalServerError("error locking rows")
	}

	var ID int64

	// Perform the upsert operation
	if err == sql.ErrNoRows {
		// Insert new payment
		queryInsert := fmt.Sprintf(`
			INSERT INTO payments (booking_id, amount, currency, status, payment_method, payment_date, payment_expiration)
			VALUES ('%s', %f, '%s', '%s', '%s', '%s', '%s') RETURNING id
		`, payment.BookingID.String(), payment.Amount, payment.Currency, payment.Status, payment.PaymentMethod, payment.PaymentDate.Format("2006-01-02 15:04:05"), payment.PaymentExpiration.Format("2006-01-02 15:04:05"))
		err := tx.QueryRowContext(ctx, queryInsert).Scan(&ID)
		if err != nil {
			fmt.Println("err msg 2", err)
			tx.Rollback()
			return errors.InternalServerError("error upserting payment")
		}
	} else {
		// Update existing payment
		queryUpdate := fmt.Sprintf(`
			UPDATE payments
			SET amount = %f, currency = '%s', status = '%s', payment_method = '%s', payment_date = '%s', payment_expiration = '%s'
			WHERE booking_id = '%s' RETURNING id
		`, payment.Amount, payment.Currency, payment.Status, payment.PaymentMethod, payment.PaymentDate.Format("2006-01-02 15:04:05"), payment.PaymentExpiration.Format("2006-01-02 15:04:05"), payment.BookingID.String())
		err := tx.QueryRowContext(ctx, queryUpdate).Scan(&ID)
		if err != nil {
			fmt.Println("err msg 3", err)
			tx.Rollback()
			return errors.InternalServerError("error upserting payment")
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("err msg 4", err)
		return errors.InternalServerError("error committing transaction")
	}

	return nil
}

// FindBookingByUserID implements Repositories.
func (r *repositories) FindBookingByUserID(ctx context.Context, userID int64) (entity.Booking, error) {
	query := fmt.Sprintf(`SELECT * FROM bookings WHERE user_id = %d`, userID)
	var booking entity.Booking
	err := r.db.Get(&booking, query)
	if err == sql.ErrNoRows {
		return entity.Booking{}, nil
	}
	if err != nil {
		return entity.Booking{}, errors.InternalServerError("error find booking by user id")
	}
	return booking, nil
}

// FindPaymentByBookingID implements Repositories.
func (r *repositories) FindPaymentByBookingID(ctx context.Context, bookingID string) (entity.Payment, error) {
	bookingIDuuid := uuid.MustParse(bookingID)
	query := fmt.Sprintf(`SELECT * FROM payments WHERE booking_id = '%s'`, bookingIDuuid.String())
	var payment entity.Payment
	err := r.db.Get(&payment, query)
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
	url := fmt.Sprintf("http://%s:%s/api/private/user/validate?token=%s", r.cfgUserService.Host, r.cfgUserService.Port, token)
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
	var respBase response.BaseResponse

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respBase); err != nil {
		return response.UserServiceValidate{
			IsValid: false,
			UserID:  0,
		}, err
	}

	respBase.Data = respBase.Data.(map[string]interface{})
	respData := response.UserServiceValidate{
		IsValid: respBase.Data.(map[string]interface{})["is_valid"].(bool),
		UserID:  int64(respBase.Data.(map[string]interface{})["user_id"].(float64)),
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
