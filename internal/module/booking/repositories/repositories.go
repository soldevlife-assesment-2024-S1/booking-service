package repositories

import (
	"booking-service/config"
	"booking-service/internal/module/booking/models/entity"
	"booking-service/internal/module/booking/models/response"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/log"
	"context"
	"encoding/json"
	"fmt"

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

// FindBookingByUserID implements Repositories.
func (r *repositories) FindBookingByUserID(ctx context.Context, userID int64) (entity.Booking, error) {
	panic("unimplemented")
}

// FindPaymentByBookingID implements Repositories.
func (r *repositories) FindPaymentByBookingID(ctx context.Context, bookingID int64) (entity.Payment, error) {
	panic("unimplemented")
}

type Repositories interface {
	// http
	ValidateToken(ctx context.Context, token string) (response.UserServiceValidate, error)
	// db
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
