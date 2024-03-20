package usecases

import (
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/models/response"
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/log"
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type usecase struct {
	repo    repositories.Repositories
	log     log.Logger
	publish message.Publisher
}

// Payment implements Usecase.
func (u *usecase) Payment(ctx context.Context, payload *request.Payment) error {
	// 1. check if payment is valid
	// 2. insert to db
	// 3. publish to rabbit mq for decrement stock ticket
	// 4. send notification to user about payment
	return nil
}

type Usecase interface {
	// http
	BookTicket(ctx context.Context, payload *request.BookTicket, userID int64) error
	ConsumeBookTicketQueue(ctx context.Context, payload *request.BookTicket) error
	ShowBookings(ctx context.Context, userID int64) (response.BookedTicket, error)
	Payment(ctx context.Context, payload *request.Payment) error
}

func New(repo repositories.Repositories, log log.Logger) Usecase {
	return &usecase{
		repo: repo,
		log:  log,
	}
}

func (u *usecase) BookTicket(ctx context.Context, payload *request.BookTicket, userID int64) error {
	// scenario 1: booking satu satu
	// TODO: check seat stock ticket
	stock, err := u.repo.CheckStockTicket(ctx, payload.TicketDetailID)
	if err != nil {
		return errors.InternalServerError("error check stock ticket")
	}

	if stock <= 0 {
		return errors.BadRequest("stock ticket is empty")
	}

	// TODO: check if user already booked more than 2 tickets

	booking, err := u.repo.FindBookingByUserID(ctx, userID)
	if err != nil {
		return errors.InternalServerError("error find booking by user id")
	}

	if booking.TotalTickets >= 2 {
		return errors.BadRequest("user already booked more than 2 tickets")
	}

	// TODO: Book ticket
	// 1. send the queue to rabbit mq

	messageUUID := watermill.NewUUID()
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.InternalServerError("error marshal payload")
	}

	u.publish.Publish("book_ticket", message.NewMessage(messageUUID, jsonPayload))

	// TODO: send notification to user that ticket has been queued

	u.publish.Publish("notification", message.NewMessage(watermill.NewUUID(), []byte("your ticket has been queued")))

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

	if bookings.ID == "" {
		return response.BookedTicket{}, errors.NotFound("booking not found")
	}

	payment, err := u.repo.FindPaymentByBookingID(ctx, userID)
	if err != nil {
		return response.BookedTicket{}, errors.InternalServerError("error find payment by booking id")
	}

	if payment.ID == 0 {
		return response.BookedTicket{}, errors.NotFound("payment not found")
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
