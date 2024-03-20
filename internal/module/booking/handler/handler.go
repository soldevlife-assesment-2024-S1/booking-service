package handler

import (
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/usecases"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/helpers"
	"booking-service/internal/pkg/log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type BookingHandler struct {
	Log       log.Logger
	Validator *validator.Validate
	Usecase   usecases.Usecase
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

	// call usecase to book ticket

	return nil
}

func (h *BookingHandler) Payment(ctx *fiber.Ctx) error {
	return ctx.SendString("Payment Ticket")
}

func (h *BookingHandler) ShowBookings(ctx *fiber.Ctx) error {
	return ctx.SendString("Show Tickets")
}
