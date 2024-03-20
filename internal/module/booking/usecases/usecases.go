package usecases

import (
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/models/response"
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/log"
	"context"
)

type usecase struct {
	repo repositories.Repositories
	log  log.Logger
}

type Usecase interface {
	// http
	BookTicket(ctx context.Context, payload *request.BookTicket) error
	ConsumeBookTicketQueue(ctx context.Context, payload *request.BookTicket) error
	ShowBookings(ctx context.Context, userID int64) (response.BookedTicket, error)
}

func New(repo repositories.Repositories, log log.Logger) Usecase {
	return &usecase{
		repo: repo,
		log:  log,
	}
}

func (u *usecase) BookTicket(ctx context.Context, payload *request.BookTicket) error {
	// scenario 1: booking satu satu

	// TODO: check seat stock ticket

	// TODO: check if user already booked more than 2 tickets

	// TODO: Book ticket
	// 1. send the queue to rabbit mq

	// TODO: send notification to user that ticket has been queued

	// scenario 2: booking banyak
	return nil
}

func (u *usecase) ConsumeBookTicketQueue(ctx context.Context, payload *request.BookTicket) error {
	// 1. check stock ticket
	// 2. decrement to redis stock ticket
	// 3. set booking expired time and payment expired time
	// 5. insert to db (lock table) or use optimistic lock
	// 6. publish to rabbit mq for decrement stock ticket to ticket service
	// 7. send notification to user about payment
	return nil
}

func (u *usecase) ShowBookings(ctx context.Context, userID int64) (response.BookedTicket, error) {
	// 1. find user booking from db
	bookings, err := u.repo.FindBookingByUserID(ctx, userID)
	if err != nil {
		return response.BookedTicket{}, errors.InternalServerError("error find booking by user id")
	}

	payment, err := u.repo.FindPaymentByBookingID(ctx, userID)
	if err != nil {
		return response.BookedTicket{}, errors.InternalServerError("error find payment by booking id")
	}

	response := response.BookedTicket{
		ID:            bookings.ID,
		FullName:      bookings.FullName,
		PersonalID:    bookings.PersonalID,
		BookingDate:   bookings.BookingDate.Format("2006-01-02 15:04:05"),
		PaymentExpiry: payment.PaymentExpiration.Format("2006-01-02 15:04:05"),
		TotalAmount:   float64(payment.Amount),
		PaymentMethod: payment.PaymentMethod,
		Status:        payment.Status,
	}
	// 3. return booking
	return response, nil
}
