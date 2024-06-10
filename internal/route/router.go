package router

import (
	"booking-service/internal/module/booking/handler"
	"booking-service/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func Initialize(app *fiber.App, handlerBooking *handler.BookingHandler, m *middleware.Middleware) *fiber.App {

	// health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("OK")
	})

	Api := app.Group("/api")

	// public routes
	v1 := Api.Group("/v1")
	v1.Get("/bookings", m.ValidateToken, handlerBooking.ShowBookings)
	// v1.Post("/book", m.ValidateToken, m.CheckIsWeekend, handlerBooking.BookTicket)
	v1.Post("/book", m.ValidateToken, handlerBooking.BookTicket)
	v1.Post("/payment", m.ValidateToken, handlerBooking.Payment)
	v1.Post("/payment/cancel", m.ValidateToken, handlerBooking.PaymentCancel)

	private := Api.Group("/private")
	private.Get("/payment/pending", handlerBooking.CountPendingPayment)

	return app

}
