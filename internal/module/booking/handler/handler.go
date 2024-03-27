package handler

import (
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/usecases"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/helpers"
	"booking-service/internal/pkg/log"
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
)

type BookingHandler struct {
	Log       log.Logger
	Validator *validator.Validate
	Usecase   usecases.Usecase
	Publish   message.Publisher
}

func (h *BookingHandler) BookTicket(ctx *fiber.Ctx) error {
	var req request.BookTicket
	if err := ctx.BodyParser(&req); err != nil {
		h.Log.Error(ctx.Context(), "error parse request", err)
		return helpers.RespError(ctx, h.Log, errors.BadRequest("error parse request"))
	}

	if err := h.Validator.Struct(req); err != nil {
		h.Log.Error(ctx.Context(), "error validate request", err)
		return helpers.RespError(ctx, h.Log, errors.BadRequest("error validate request"))
	}

	userID := ctx.Locals("user_id").(int64)

	// call usecase to book ticket
	err := h.Usecase.BookTicket(ctx.Context(), &req, userID)
	if err != nil {
		h.Log.Error(ctx.Context(), "error book ticket", err)
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
		h.Log.Error(msg.Context(), "error unmarshal message", err)

		// publish to poison queue
		reqPoisoned := request.PoisonedQueue{
			TopicTarget: "book_ticket",
			ErrorMsg:    err.Error(),
			Payload:     msg.Payload,
		}

		jsonPayload, _ := json.Marshal(reqPoisoned)

		h.Publish.Publish("poisoned_queue", message.NewMessage(watermill.NewUUID(), jsonPayload))

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
		h.Publish.Publish("poisoned_queue", message.NewMessage(watermill.NewUUID(), jsonPayload))

		h.Log.Error(msg.Context(), "error consume booking queue", err)

		return err
	}

	return nil
}

func (h *BookingHandler) Payment(ctx *fiber.Ctx) error {
	var req request.Payment
	if err := ctx.BodyParser(&req); err != nil {
		h.Log.Error(ctx.Context(), "error parse request", err)
		return helpers.RespError(ctx, h.Log, errors.BadRequest("error parse request"))
	}

	if err := h.Validator.Struct(req); err != nil {
		h.Log.Error(ctx.Context(), "error validate request", err)
		return helpers.RespError(ctx, h.Log, errors.BadRequest(err.Error()))
	}

	// call usecase to payment
	err := h.Usecase.Payment(ctx.Context(), &req)
	if err != nil {
		h.Log.Error(ctx.Context(), "error payment", err)
		return helpers.RespError(ctx, h.Log, err)
	}

	return helpers.RespSuccess(ctx, h.Log, nil, "success payment")
}

func (h *BookingHandler) ShowBookings(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(int64)

	// call usecase to show bookings
	resp, err := h.Usecase.ShowBookings(ctx.Context(), userID)
	if err != nil {
		h.Log.Error(ctx.Context(), "error show bookings", err)
		return helpers.RespError(ctx, h.Log, err)
	}

	return helpers.RespSuccess(ctx, h.Log, resp, "success show bookings")
}

func (h *BookingHandler) SetPaymentExpired(ctx context.Context, t *asynq.Task) error {
	var req request.PaymentExpiration
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		h.Log.Error(ctx, "error unmarshal message", err)
		return err
	}

	if err := h.Validator.Struct(req); err != nil {
		h.Log.Error(ctx, "error validate request", err)
		return err
	}

	// call usecase to set payment expired
	err := h.Usecase.SetPaymentExpired(ctx, &req)
	if err != nil {
		h.Log.Error(ctx, "error set payment expired", err)
		return err
	}

	return nil
}
