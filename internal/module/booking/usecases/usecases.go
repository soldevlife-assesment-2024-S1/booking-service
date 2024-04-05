package usecases

import (
	"booking-service/internal/module/booking/models/entity"
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/models/response"
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/helpers"
	"booking-service/internal/pkg/log"
	"booking-service/internal/pkg/scheduler"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type usecase struct {
	repo            repositories.Repositories
	log             log.Logger
	publish         message.Publisher
	clientScheduler *asynq.Client
}

// PaymentCancel implements Usecase.
func (u *usecase) PaymentCancel(ctx context.Context, payload *request.PaymentCancellation) error {
	// 1. find payment by booking id
	payment, err := u.repo.FindPaymentByBookingID(ctx, payload.BookingID)
	if err != nil {
		return errors.InternalServerError("error find payment by booking id")
	}

	// 2. find booking by booking id
	booking, err := u.repo.FindBookingByID(ctx, payment.BookingID.String())
	if err != nil {
		return errors.InternalServerError("error find booking by booking id")
	}

	if payment.ID == 0 {
		return errors.NotFound("payment not found")
	}

	// 2. if payment status is pending
	if payment.Status == "pending" {
		// 3. update payment status to expired
		payment.Status = "cancelled"
		err = u.repo.UpsertPayment(ctx, &payment)
		if err != nil {
			return errors.InternalServerError("error upsert payment")
		}

		// 4. cancel job to check payment expired time

		err = u.repo.DeleteTaskScheduler(ctx, payment.TaskID)
		if err != nil {
			return errors.InternalServerError("error delete task scheduler")
		}

		// 5. publish to rabbit mq for increment stock ticket to ticket service

		messageUUID := watermill.NewUUID()

		specPayload := request.DecrementStockTicket{
			TicketDetailID: booking.TicketDetailID,
			TotalTickets:   booking.TotalTickets,
		}

		jsonPayload, err := json.Marshal(specPayload)
		if err != nil {
			return errors.InternalServerError("error marshal payload")
		}

		err = u.repo.IncrementStockTicket(ctx, booking.TicketDetailID)
		if err != nil {
			return errors.InternalServerError("error increment stock ticket")
		}

		err = u.publish.Publish("increment_stock_ticket", message.NewMessage(messageUUID, jsonPayload))
		if err != nil {
			return errors.InternalServerError("error publish decrement stock ticket")
		}

		// 6. send notification to user about payment

		payloadNotification := request.NotificationMessage{
			Message: "your payment has been cancelled",
		}

		jsonPayloadNotification, err := json.Marshal(payloadNotification)
		if err != nil {
			return errors.InternalServerError("error marshal payload")
		}

		err = u.publish.Publish("notification_cancel", message.NewMessage(watermill.NewUUID(), jsonPayloadNotification))
		if err != nil {
			return errors.InternalServerError("error publish notification")
		}
	}

	return nil
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

	bookingID := uuid.MustParse(payload.BookingID)

	specPayment := entity.Payment{
		BookingID:         bookingID,
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

	// cancel job to check payment expired time

	err = u.repo.DeleteTaskScheduler(ctx, dataPayment.TaskID)
	if err != nil {
		return errors.InternalServerError("error delete task scheduler")
	}

	// 4. send notification to user about payment

	payloadNotification := request.NotificationPayment{
		BookingID:     payload.BookingID,
		Message:       "your payment has been paid",
		PaymentMethod: payload.PaymetMethod,
	}

	jsonPayloadNotification, err := json.Marshal(payloadNotification)
	if err != nil {
		return errors.InternalServerError("error marshal payload")
	}

	err = u.publish.Publish("notification_payment", message.NewMessage(watermill.NewUUID(), jsonPayloadNotification))
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
	PaymentCancel(ctx context.Context, payload *request.PaymentCancellation) error
	SetPaymentExpired(ctx context.Context, payload *request.PaymentExpiration) error
}

func New(repo repositories.Repositories, log log.Logger, publish message.Publisher, clientScheduler *asynq.Client) Usecase {
	return &usecase{
		repo:            repo,
		log:             log,
		publish:         publish,
		clientScheduler: clientScheduler,
	}
}

func (u *usecase) BookTicket(ctx context.Context, payload *request.BookTicket, userID int64) error {
	// scenario 1: booking satu satu
	stock, err := u.repo.CheckStockTicket(ctx, payload.TicketDetailID)
	if err != nil {
		return errors.InternalServerError("error check stock ticket")
	}

	if stock <= 0 {
		return errors.BadRequest("stock ticket is empty")
	}

	booking, err := u.repo.FindBookingByUserID(ctx, userID)
	if err != nil {
		return errors.InternalServerError("error find booking by user id")
	}

	if booking.TotalTickets >= 2 {
		return errors.BadRequest("user already booked more than 2 tickets")
	}

	// 1. send the queue to rabbit mq

	messageUUID := watermill.NewUUID()
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.InternalServerError("error marshal payload")
	}

	u.publish.Publish("book_ticket", message.NewMessage(messageUUID, jsonPayload))

	payloadNotification := request.NotificationMessage{
		Message: "your ticket has been queued",
	}

	jsonPayloadNotification, err := json.Marshal(payloadNotification)
	if err != nil {
		return errors.InternalServerError("error marshal payload")
	}

	u.publish.Publish("notification_queue", message.NewMessage(watermill.NewUUID(), jsonPayloadNotification))

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

	paymentExpiredAt := time.Now().Add(time.Minute * 30)

	// 5. insert to db (lock table) or use optimistic lock

	specBooking := entity.Booking{
		UserID:         payload.UserID,
		TicketDetailID: payload.TicketDetailID,
		TotalTickets:   payload.TotalTickets,
		FullName:       payload.FullName,
		PersonalID:     payload.PersonalID,
		BookingDate:    time.Now(),
	}

	bookingID, err := u.repo.UpsertBooking(ctx, &specBooking)
	if err != nil {
		fmt.Println(err)
		return errors.InternalServerError("error upsert booking")
	}

	// request to calculate total amount to ticket service

	amount, err := u.repo.InquiryTicketAmount(ctx, payload.TicketDetailID, payload.TotalTickets)
	if err != nil {
		return errors.InternalServerError("error inquiry ticket amount")
	}

	// 6. start job to check payment expired time

	specPaymentExpiration := request.PaymentExpiration{
		BookingID:      bookingID,
		TicketDetailID: payload.TicketDetailID,
		TotalTickets:   payload.TotalTickets,
	}

	jsonPayloadScheduler, err := json.Marshal(specPaymentExpiration)
	if err != nil {
		return errors.InternalServerError("error marshal payload")
	}

	expiredAt := helpers.DurationCalculation(paymentExpiredAt)

	taskPaymentExpiredAt := asynq.NewTask(scheduler.TypeSetPaymentExpired, jsonPayloadScheduler, asynq.MaxRetry(3), asynq.Timeout(expiredAt))

	taskInfo, err := u.clientScheduler.Enqueue(taskPaymentExpiredAt, asynq.ProcessIn(expiredAt))
	if err != nil {
		return errors.InternalServerError("error enqueue task payment expired")
	}

	bookingIDuuid := uuid.MustParse(bookingID)

	specPayment := entity.Payment{
		BookingID:         bookingIDuuid,
		Amount:            amount,
		Currency:          "IDR",
		Status:            "pending",
		PaymentMethod:     "",
		TaskID:            taskInfo.ID,
		PaymentExpiration: paymentExpiredAt,
	}

	err = u.repo.UpsertPayment(ctx, &specPayment)
	if err != nil {
		return errors.InternalServerError("error upsert payment")
	}

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

	payloadNotification := request.NotificationInvoice{
		BookingID:         bookingID,
		PaymentAmount:     amount,
		PaymentExpiration: paymentExpiredAt.Format("2006-01-02 15:04:05"),
	}

	jsonPayloadNotification, err := json.Marshal(payloadNotification)
	if err != nil {
		return errors.InternalServerError("error marshal payload")
	}

	err = u.publish.Publish("notification_invoice", message.NewMessage(watermill.NewUUID(), jsonPayloadNotification))
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

	if bookings.ID.String() == "" {
		return response.BookedTicket{}, errors.NotFound("booking not found")
	}

	payment, err := u.repo.FindPaymentByBookingID(ctx, bookings.ID.String())
	if err != nil {
		return response.BookedTicket{}, errors.InternalServerError("error find payment by booking id")
	}

	if payment.ID == 0 {
		return response.BookedTicket{}, errors.NotFound("payment not found")
	}

	response := response.BookedTicket{
		ID:            bookings.ID.String(),
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

func (u *usecase) SetPaymentExpired(ctx context.Context, payload *request.PaymentExpiration) error {
	// 1. find payment by booking id
	payment, err := u.repo.FindPaymentByBookingID(ctx, payload.BookingID)
	if err != nil {
		u.log.Error(ctx, "error find payment by booking id", err)
		return err
	}

	// 2. if payment status is pending and payment expired time is now
	if payment.Status == "pending" {
		// 3. update payment status to expired
		payment.Status = "expired"
		err = u.repo.UpsertPayment(ctx, &payment)
		if err != nil {
			u.log.Error(ctx, "error upsert payment", err)
			return err
		}

		// 4. publish to rabbit mq for increment stock ticket to ticket service

		messageUUID := watermill.NewUUID()

		specPayload := request.DecrementStockTicket{
			TicketDetailID: payload.TicketDetailID,
			TotalTickets:   payload.TotalTickets,
		}

		jsonPayload, err := json.Marshal(specPayload)
		if err != nil {
			return errors.InternalServerError("error marshal payload")
		}

		err = u.repo.IncrementStockTicket(ctx, payload.TicketDetailID)
		if err != nil {
			return errors.InternalServerError("error increment stock ticket")
		}

		err = u.publish.Publish("increment_stock_ticket", message.NewMessage(messageUUID, jsonPayload))

		if err != nil {
			u.log.Error(ctx, "error publish decrement stock ticket", err)
			return err
		}
	}

	return nil
}
