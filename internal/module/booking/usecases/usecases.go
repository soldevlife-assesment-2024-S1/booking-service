package usecases

import (
	"booking-service/internal/module/booking/models/entity"
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/models/response"
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/log"
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/reugn/go-quartz/quartz"
)

type usecase struct {
	repo     repositories.Repositories
	log      log.Logger
	publish  message.Publisher
	jobQueue quartz.JobQueue
}

// Payment implements Usecase.
func (u *usecase) Payment(ctx context.Context, payload *request.Payment) error {
	// 1. check if payment is valid

	dataPayment, err := u.repo.FindPaymentByBookingID(ctx, payload.BookingID)
	if err != nil {
		return errors.InternalServerError("error repository find payment by booking id")
	}

	if dataPayment.ID == 0 {
		return errors.NotFound("payment not found")
	}

	if dataPayment.Status != "pending" {
		return errors.BadRequest("payment already paid / expired")
	}

	// 2. insert to db

	specPayment := entity.Payment{
		BookingID:         payload.BookingID,
		Amount:            payload.TotalAmount,
		Currency:          "USD",
		Status:            "paid",
		PaymentMethod:     payload.PaymetMethod,
		PaymentDate:       time.Now(),
		PaymentExpiration: dataPayment.PaymentExpiration,
	}

	err = u.repo.UpsertPayment(ctx, &specPayment)
	if err != nil {
		return errors.InternalServerError("error upsert payment")
	}

	// 3. publish to rabbit mq for decrement stock ticket

	dataBooking, err := u.repo.FindBookingByBookingID(ctx, payload.BookingID)
	if err != nil {
		return errors.InternalServerError("error find booking by booking id")
	}

	messageUUID := watermill.NewUUID()

	specPayload := request.DecrementStockTicket{
		TicketDetailID: dataBooking.TicketDetailID,
		TotalTickets:   1,
	}

	jsonPayload, err := json.Marshal(specPayload)
	if err != nil {
		return errors.InternalServerError("error marshal payload")
	}

	err = u.publish.Publish("decrement_stock_ticket", message.NewMessage(messageUUID, jsonPayload))
	if err != nil {
		return errors.InternalServerError("error publish decrement stock ticket")
	}

	// 4. send notification to user about payment

	err = u.publish.Publish("notification", message.NewMessage(watermill.NewUUID(), []byte("your payment has been paid")))
	if err != nil {
		return errors.InternalServerError("error publish notification")
	}

	return nil
}

type Usecase interface {
	// http
	BookTicket(ctx context.Context, payload *request.BookTicket, userID int64) error
	ConsumeBookTicketQueue(ctx context.Context, payload *request.BookTicket) error
	ShowBookings(ctx context.Context, userID int64) (response.BookedTicket, error)
	Payment(ctx context.Context, payload *request.Payment) error
}

func New(repo repositories.Repositories, log log.Logger, publish message.Publisher) Usecase {
	return &usecase{
		repo:    repo,
		log:     log,
		publish: publish,
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

	stock, err := u.repo.CheckStockTicket(ctx, payload.TicketDetailID)
	if err != nil {
		return errors.InternalServerError("error check stock ticket")
	}

	if stock <= 0 {
		return errors.BadRequest("stock ticket is empty")
	}

	// 2. decrement to redis stock ticket

	err = u.repo.DecrementStockTicket(ctx, payload.TicketDetailID)
	if err != nil {
		return errors.InternalServerError("error decrement stock ticket")
	}

	// 3. set booking expired time and payment expired time

	bookExpiredAt := time.Now().Add(time.Hour * 24 * 3)
	paymentExpiredAt := time.Now().Add(time.Hour * 24 * 1)

	// 5. insert to db (lock table) or use optimistic lock

	specBooking := entity.Booking{
		UserID:            payload.UserID,
		TicketDetailID:    payload.TicketDetailID,
		TotalTickets:      payload.TotalTickets,
		FullName:          payload.FullName,
		PersonalID:        payload.PersonalID,
		BookingDate:       time.Now(),
		BookingExpiration: bookExpiredAt,
	}

	bookingID, err := u.repo.UpsertBooking(ctx, &specBooking)
	if err != nil {
		return errors.InternalServerError("error upsert booking")
	}

	// request to calculate total amount to ticket service

	specPayment := entity.Payment{
		BookingID:         bookingID,
		Amount:            0,
		Currency:          "IDR",
		Status:            "pending",
		PaymentMethod:     "",
		PaymentDate:       time.Now(),
		PaymentExpiration: paymentExpiredAt,
	}

	err = u.repo.UpsertPayment(ctx, &specPayment)
	if err != nil {
		return errors.InternalServerError("error upsert payment")
	}

	// 6. start job to check payment expired time

	// scheduledJob := quartz.NewJobDetail(func(ctx context.Context) {
	// 	// 1. find payment by booking id
	// 	payment, err := u.repo.FindPaymentByBookingID(ctx, specPayment.ID)
	// 	if err != nil {
	// 		u.log.Error(ctx, "error find payment by booking id", err)
	// 	}

	// 	// 2. if payment status is pending and payment expired time is now
	// 	if payment.Status == "pending" && payment.PaymentExpiration.Before(time.Now()) {
	// 		// 3. update payment status to expired
	// 		payment.Status = "expired"
	// 		err = u.repo.UpsertPayment(ctx, &payment)
	// 		if err != nil {
	// 			u.log.Error(ctx, "error upsert payment", err)
	// 		}
	// 	}
	// }, quartz.NewJobKey("check_payment_expired_time"))

	// 7. publish to rabbit mq for decrement stock ticket to ticket service

	messageUUID := watermill.NewUUID()

	specPayload := request.DecrementStockTicket{
		TicketDetailID: payload.TicketDetailID,
		TotalTickets:   payload.TotalTickets,
	}

	jsonPayload, err := json.Marshal(specPayload)
	if err != nil {
		return errors.InternalServerError("error marshal payload")
	}

	err = u.publish.Publish("decrement_stock_ticket", message.NewMessage(messageUUID, jsonPayload))
	if err != nil {
		u.log.Error(ctx, "error publish decrement stock ticket", err)
	}

	// 8. send notification to user about payment

	err = u.publish.Publish("notification", message.NewMessage(watermill.NewUUID(), []byte("your ticket has been queued")))
	if err != nil {
		u.log.Error(ctx, "error publish notification", err)
	}

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

	payment, err := u.repo.FindPaymentByBookingID(ctx, bookings.ID)
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
