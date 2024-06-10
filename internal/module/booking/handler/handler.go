package handler

import (
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/usecases"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/helpers"
	"context"
	"fmt"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type BookingHandler struct {
	Log       *otelzap.Logger
	Validator *validator.Validate
	Usecase   usecases.Usecase
	Publish   message.Publisher
}

func (h *BookingHandler) BookTicket(ctx *fiber.Ctx) error {
	var req request.BookTicket
	if err := ctx.BodyParser(&req); err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error parse request: %v", err))
		return helpers.RespError(ctx, h.Log, errors.BadRequest("error parse request"))
	}

	if err := h.Validator.Struct(req); err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error validate request: %v", err))
		return helpers.RespError(ctx, h.Log, errors.BadRequest("error validate request"))
	}

	userID := ctx.Locals("user_id").(int64)
	emailUser := ctx.Locals("email_user").(string)

	// call usecase to book ticket
	err := h.Usecase.BookTicket(ctx.UserContext(), &req, userID, emailUser)
	if err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error book ticket: %v", err))
		return helpers.RespError(ctx, h.Log, err)
	}

	resp := map[string]interface{}{
		"message": "success book ticket, please check your email for payment ticket",
	}

	return helpers.RespSuccess(ctx, h.Log, resp, "success book ticket, please check your email for payment ticket")
}

func (h *BookingHandler) ConsumeBookingQueue(msg *message.Message) error {
	msg.Ack() // acknowledge message
	var req request.BookTicket
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		h.Log.Ctx(msg.Context()).Error(fmt.Sprintf("error unmarshal message: %v", err))

		// publish to poison queue
		reqPoisoned := request.PoisonedQueue{
			TopicTarget: "book_ticket",
			ErrorMsg:    err.Error(),
			Payload:     msg.Payload,
		}

		jsonPayload, _ := json.Marshal(reqPoisoned)

		err = h.Publish.Publish("poisoned_queue", message.NewMessage(watermill.NewUUID(), jsonPayload))
		if err != nil {
			h.Log.Ctx(msg.Context()).Error(fmt.Sprintf("error publish to poison queue: %v", err))
		}

		return err
	}

	ctx := context.Background()

	// call usecase to consume booking queue
	err := h.Usecase.ConsumeBookTicketQueue(ctx, &req)
	if err != nil {
		// publish to poison queue
		reqPoisoned := request.PoisonedQueue{
			TopicTarget: "book_ticket",
			ErrorMsg:    err.Error(),
			Payload:     msg.Payload,
		}

		jsonPayload, _ := json.Marshal(reqPoisoned)
		err = h.Publish.Publish("poisoned_queue", message.NewMessage(watermill.NewUUID(), jsonPayload))
		if err != nil {
			h.Log.Ctx(msg.Context()).Error(fmt.Sprintf("error publish to poison queue: %v", err))
		}

		h.Log.Ctx(msg.Context()).Error(fmt.Sprintf("error consume booking queue: %v", err))

		return err
	}

	return nil
}

func (h *BookingHandler) Payment(ctx *fiber.Ctx) error {
	var req request.Payment
	if err := ctx.BodyParser(&req); err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error parse request: %v", err))
		return helpers.RespError(ctx, h.Log, errors.BadRequest("error parse request"))
	}

	if err := h.Validator.Struct(req); err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error validate request: %v", err))
		return helpers.RespError(ctx, h.Log, errors.BadRequest(err.Error()))
	}

	emailUser := ctx.Locals("email_user").(string)

	// call usecase to payment
	err := h.Usecase.Payment(ctx.UserContext(), &req, emailUser)
	if err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error payment: %v", err))
		return helpers.RespError(ctx, h.Log, err)
	}

	return helpers.RespSuccess(ctx, h.Log, nil, "success payment")
}

func (h *BookingHandler) PaymentCancel(ctx *fiber.Ctx) error {
	var req request.PaymentCancellation
	if err := ctx.BodyParser(&req); err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error parse request: %v", err))
		return helpers.RespError(ctx, h.Log, errors.BadRequest("error parse request"))
	}

	if err := h.Validator.Struct(req); err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error validate request: %v", err))
		return helpers.RespError(ctx, h.Log, errors.BadRequest(err.Error()))
	}

	emailUser := ctx.Locals("email_user").(string)

	// call usecase to payment cancel
	err := h.Usecase.PaymentCancel(ctx.UserContext(), &req, emailUser)
	if err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error payment cancel: %v", err))
		return helpers.RespError(ctx, h.Log, err)
	}

	return helpers.RespSuccess(ctx, h.Log, nil, "success payment cancel")
}

func (h *BookingHandler) ShowBookings(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(int64)

	// call usecase to show bookings
	resp, err := h.Usecase.ShowBookings(ctx.UserContext(), userID)
	if err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error show bookings: %v", err))
		return helpers.RespError(ctx, h.Log, err)
	}

	return helpers.RespSuccess(ctx, h.Log, resp, "success show bookings")
}

func (h *BookingHandler) CountPendingPayment(ctx *fiber.Ctx) error {
	TicketDetailID := ctx.Query("ticket_detail")
	ticketDetailIDInt64, err := strconv.ParseInt(TicketDetailID, 10, 64)
	if err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error parse ticket detail id: %v", err))
		return helpers.RespError(ctx, h.Log, errors.BadRequest("error parse ticket detail id"))
	}
	// call usecase to count pending payment
	resp, err := h.Usecase.CountPendingPayment(ctx.UserContext(), ticketDetailIDInt64)
	if err != nil {
		h.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error count pending payment: %v", err))
		return helpers.RespError(ctx, h.Log, err)
	}

	return helpers.RespSuccess(ctx, h.Log, resp, "success count pending payment")
}

func (h *BookingHandler) SetPaymentExpired(ctx context.Context, t *asynq.Task) error {
	var req request.PaymentExpiration
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		h.Log.Ctx(ctx).Error(fmt.Sprintf("error unmarshal payload: %v", err))
		return err
	}

	if err := h.Validator.Struct(req); err != nil {
		h.Log.Ctx(ctx).Error(fmt.Sprintf("error validate payload: %v", err))
		return err
	}

	// call usecase to set payment expired
	err := h.Usecase.SetPaymentExpired(ctx, &req)
	if err != nil {
		h.Log.Ctx(ctx).Error(fmt.Sprintf("error set payment expired: %v", err))
		return err
	}

	return nil
}
