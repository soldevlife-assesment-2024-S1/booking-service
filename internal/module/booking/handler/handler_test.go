package handler_test

import (
	"booking-service/internal/module/booking/handler"
	"booking-service/internal/module/booking/mocks"
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/models/response"
	log_internal "booking-service/internal/pkg/log"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/valyala/fasthttp"
)

var (
	h             *handler.BookingHandler
	ucm           *mocks.Usecase
	logMock       *otelzap.Logger
	app           *fiber.App
	validatorTest *validator.Validate
	p             message.Publisher
	asyncTask     *asynq.Task
)

type mockPublisher struct{}

// Close implements message.Publisher.
func (m *mockPublisher) Close() error {
	return nil
}

// Publish implements message.Publisher.
func (m *mockPublisher) Publish(topic string, messages ...*message.Message) error {
	return nil
}

func NewMockPublisher() message.Publisher {
	return &mockPublisher{}
}

func setup() {
	ucm = &mocks.Usecase{}
	logMock := log_internal.Setup()
	validatorTest = validator.New()
	p = NewMockPublisher()
	h = &handler.BookingHandler{
		Log:       logMock,
		Validator: validatorTest,
		Usecase:   ucm,
		Publish:   p,
	}
	app = fiber.New()
}

func teardown() {
	ucm = nil
	logMock = nil
	validatorTest = nil
	p = nil
	h = nil
	app = nil
}

func TestBookTicket(t *testing.T) {
	setup()
	defer teardown()
	t.Run("Test case 1 | BookTicket", func(t *testing.T) {
		// mock data
		payload := request.BookTicket{
			TicketDetailID: 1,
			FullName:       "test",
			PersonalID:     "123",
			UserID:         1,
			TotalTickets:   1,
			EmailRecipient: "test@test.com",
		}

		jsonData, _ := json.Marshal(payload)

		httpReq := httptest.NewRequest("POST", "/api/v1/book", nil)
		httpReq.Header.Set("Content-Type", "application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetRequestURI("/api/v1/book")
		ctx.Request().Header.SetContentType("application/json")
		ctx.Request().Header.SetMethod("POST")
		ctx.Request().SetBody(jsonData)
		ctx.Locals("user_id", int64(1))
		ctx.Locals("email_user", "test@test.com")

		// mock usecase
		ucm.On("BookTicket", ctx.Context(), &payload, int64(1), "test@test.com").Return(nil)

		// test
		err := h.BookTicket(ctx)

		// assertion
		assert.NoError(t, err)
	})
}

func TestConsumeBookTicketQueue(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		// mock data
		payload := request.BookTicket{
			TicketDetailID: 1,
			FullName:       "test",
			PersonalID:     "123",
			UserID:         1,
			TotalTickets:   1,
			EmailRecipient: "test@test.com",
		}

		jsonData, _ := json.Marshal(payload)

		msg := message.NewMessage("123", jsonData)

		// mock usecase
		ucm.On("ConsumeBookTicketQueue", ctx, &payload).Return(nil)

		// test
		err := h.ConsumeBookingQueue(msg)

		// assertion
		assert.NoError(t, err)
	})
}

func TestPayment(t *testing.T) {
	setup()
	defer teardown()
	t.Run("Success", func(t *testing.T) {
		// mock data
		payload := request.Payment{
			BookingID:    "123",
			TotalAmount:  1000,
			PaymetMethod: "ovo",
		}

		jsonData, _ := json.Marshal(payload)

		httpReq := httptest.NewRequest("POST", "/api/v1/payment", nil)
		httpReq.Header.Set("Content-Type", "application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetRequestURI("/api/v1/payment")
		ctx.Request().Header.SetContentType("application/json")
		ctx.Request().Header.SetMethod("POST")
		ctx.Request().SetBody(jsonData)
		ctx.Locals("email_user", "test@test.com")

		// mock usecase
		ucm.On("Payment", ctx.Context(), &payload, "test@test.com").Return(nil)

		// test
		err := h.Payment(ctx)

		// assertion
		assert.NoError(t, err)

	})
}

func TestPaymentCancel(t *testing.T) {
	setup()
	defer teardown()
	t.Run("Success", func(t *testing.T) {
		// mock data
		payload := request.PaymentCancellation{
			BookingID: "123",
		}

		jsonData, _ := json.Marshal(payload)

		httpReq := httptest.NewRequest("POST", "/api/v1/payment/cancel", nil)
		httpReq.Header.Set("Content-Type", "application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetRequestURI("/api/v1/payment/cancel")
		ctx.Request().Header.SetContentType("application/json")
		ctx.Request().Header.SetMethod("POST")
		ctx.Request().SetBody(jsonData)
		ctx.Locals("email_user", "test@test.com")

		// mock usecase
		ucm.On("PaymentCancel", ctx.Context(), &payload, "test@test.com").Return(nil)

		// test
		err := h.PaymentCancel(ctx)

		// assertion
		assert.NoError(t, err)

	})
}

func TestShowBooking(t *testing.T) {
	setup()
	defer teardown()
	t.Run("success", func(t *testing.T) {
		// mock data
		httpReq := httptest.NewRequest("GET", "/api/v1/bookings", nil)
		httpReq.Header.Set("Content-Type", "application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		ctx.Request().SetRequestURI("/api/v1/bookings")
		ctx.Request().Header.SetContentType("application/json")
		ctx.Request().Header.SetMethod("GET")
		ctx.Locals("user_id", int64(1))

		// mock usecase
		ucm.On("ShowBookings", ctx.Context(), int64(1)).Return(response.BookedTicket{}, nil)

		// test
		err := h.ShowBookings(ctx)

		// assertion
		assert.NoError(t, err)
	})
}

func TestSetPaymentExpired(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		// mock data
		payload := request.PaymentExpiration{
			BookingID:      "123",
			TicketDetailID: 1,
			TotalTickets:   1,
		}

		// mock usecase
		ucm.On("SetPaymentExpired", ctx, &payload).Return(nil)
		asyncTask = asynq.NewTask("set_payment_expired", []byte(`{"booking_id":"123","ticket_detail_id":1,"total_tickets":1}`))

		// test
		err := h.SetPaymentExpired(ctx, asyncTask)

		// assertion
		assert.NoError(t, err)
	})
}
