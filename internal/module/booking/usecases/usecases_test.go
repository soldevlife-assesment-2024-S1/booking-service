package usecases_test

import (
	"booking-service/internal/module/booking/mocks"
	"booking-service/internal/module/booking/models/entity"
	"booking-service/internal/module/booking/models/request"
	"booking-service/internal/module/booking/models/response"
	"booking-service/internal/module/booking/usecases"
	"booking-service/internal/pkg/helpers"
	"booking-service/internal/pkg/log"
	log_internal "booking-service/internal/pkg/log"
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	uc          usecases.Usecase
	repoMock    *mocks.Repositories
	logMock     log.Logger
	p           message.Publisher
	dateTimeNow = time.Now()
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
	repoMock = new(mocks.Repositories)
	p = NewMockPublisher()
	logZap := log_internal.SetupLogger()
	log_internal.Init(logZap)
	logMock = log_internal.GetLogger()
	uc = usecases.New(repoMock, logMock, p)
}

func teardown() {
	repoMock = nil
	uc = nil
}

func TestPaymentCancel(t *testing.T) {
	setup()
	defer teardown()

	t.Run("success", func(t *testing.T) {
		// mock data
		ctx := context.Background()
		payloadMock := request.PaymentCancellation{
			BookingID: "00000000-0000-0000-0000-000000000000",
		}
		emailUser := "teest@test.com"
		paymentMock := entity.Payment{
			ID:                1,
			BookingID:         uuid.UUID{},
			Amount:            1000,
			Currency:          "IDR",
			Status:            "pending",
			PaymentMethod:     "paypal",
			PaymentDate:       dateTimeNow,
			PaymentExpiration: time.Time{},
			TaskID:            "1",
			CreatedAt:         time.Time{},
			UpdatedAt:         sql.NullTime{},
			DeletedAt:         sql.NullTime{},
		}

		bookingMock := entity.Booking{
			ID:             uuid.UUID{},
			UserID:         1,
			TicketDetailID: 1,
			TotalTickets:   1,
			FullName:       "test",
			PersonalID:     "123",
			BookingDate:    dateTimeNow,
		}
		paymentMockUpsert := entity.Payment{
			ID:                1,
			BookingID:         uuid.UUID{},
			Amount:            1000,
			Currency:          "IDR",
			Status:            "cancelled",
			PaymentMethod:     "paypal",
			PaymentDate:       dateTimeNow,
			PaymentExpiration: time.Time{},
			TaskID:            "1",
			CreatedAt:         time.Time{},
			UpdatedAt:         sql.NullTime{},
			DeletedAt:         sql.NullTime{},
		}

		// mock repo
		repoMock.On("FindPaymentByBookingID", ctx, payloadMock.BookingID).Return(paymentMock, nil)
		repoMock.On("FindBookingByID", ctx, payloadMock.BookingID).Return(bookingMock, nil)
		repoMock.On("UpsertPayment", ctx, &paymentMockUpsert).Return(nil)
		repoMock.On("DeleteTaskScheduler", ctx, paymentMock.TaskID).Return(nil)
		repoMock.On("IncrementStockTicket", ctx, int64(1)).Return(nil)

		// test
		err := uc.PaymentCancel(ctx, &payloadMock, emailUser)
		assert.NoError(t, err)
	})
}

func TestPayment(t *testing.T) {
	setup()
	defer teardown()

	t.Run("success", func(t *testing.T) {
		// mock data
		ctx := context.Background()
		payloadMock := request.Payment{
			BookingID:    "00000000-0000-0000-0000-000000000000",
			TotalAmount:  1000,
			PaymetMethod: "Paypal",
		}
		emailUser := "test@test.com"

		paymentMock := entity.Payment{
			ID:                1,
			BookingID:         uuid.UUID{},
			Amount:            1000,
			Currency:          "USD",
			Status:            "pending",
			PaymentMethod:     "Paypal",
			PaymentDate:       dateTimeNow.Round(time.Second),
			PaymentExpiration: time.Time{},
			TaskID:            "1",
			CreatedAt:         time.Time{},
			UpdatedAt:         sql.NullTime{},
			DeletedAt:         sql.NullTime{},
		}
		paymentMockUpsert := entity.Payment{
			BookingID:         uuid.UUID{},
			Amount:            1000,
			Currency:          "USD",
			Status:            "paid",
			PaymentMethod:     "Paypal",
			PaymentDate:       dateTimeNow.Round(time.Second),
			PaymentExpiration: time.Time{},
			TaskID:            "1",
			CreatedAt:         time.Time{},
			UpdatedAt:         sql.NullTime{},
			DeletedAt:         sql.NullTime{},
		}
		// mock repo
		repoMock.On("FindPaymentByBookingID", ctx, payloadMock.BookingID).Return(paymentMock, nil)
		repoMock.On("UpsertPayment", ctx, &paymentMockUpsert).Return(nil)
		repoMock.On("DeleteTaskScheduler", ctx, paymentMock.TaskID).Return(nil)

		// test
		err := uc.Payment(ctx, &payloadMock, emailUser)

		// assert
		assert.NoError(t, err)
	})
}

func TestBookTicket(t *testing.T) {
	setup()

	defer teardown()
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		// mock data
		mockPayload := request.BookTicket{
			UserID:         1,
			TicketDetailID: 1,
			TotalTickets:   1,
			FullName:       "test",
			PersonalID:     "123",
			EmailRecipient: "test@test.com",
		}
		mockBooking := entity.Booking{
			ID:             uuid.UUID{},
			UserID:         1,
			TicketDetailID: 1,
			TotalTickets:   1,
			FullName:       "test",
			PersonalID:     "123",
			BookingDate:    dateTimeNow,
		}

		// mock repo
		repoMock.On("CheckStockTicket", ctx, mockPayload.TicketDetailID).Return(int64(100), nil)
		repoMock.On("FindBookingByUserID", ctx, mockBooking.UserID).Return(mockBooking, nil)

		// test
		err := uc.BookTicket(context.Background(), &mockPayload, mockBooking.UserID, mockPayload.EmailRecipient)

		// assert
		assert.NoError(t, err)

	})
}

func TestConsumeBookTicketQueue(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// mock data
		mockPayload := request.BookTicket{
			UserID:         1,
			TicketDetailID: 1,
			TotalTickets:   1,
			FullName:       "test",
			PersonalID:     "123",
			EmailRecipient: "test@test.com",
		}
		mockBooking := entity.Booking{
			ID:             uuid.UUID{},
			UserID:         1,
			TicketDetailID: 1,
			TotalTickets:   1,
			FullName:       "test",
			PersonalID:     "123",
			BookingDate:    dateTimeNow.Round(time.Second),
		}

		specPaymentExpiration := request.PaymentExpiration{
			BookingID:      mockBooking.ID.String(),
			TicketDetailID: mockPayload.TicketDetailID,
			TotalTickets:   mockPayload.TotalTickets,
		}

		jsonPayloadScheduler, _ := json.Marshal(specPaymentExpiration)
		paymentExpiredAt := time.Now().Add(time.Minute * 30)
		expiredAt := helpers.DurationCalculation(paymentExpiredAt)

		paymentMockUpsert := entity.Payment{
			BookingID:         uuid.UUID{},
			Amount:            100,
			Currency:          "IDR",
			Status:            "pending",
			PaymentMethod:     "",
			PaymentExpiration: paymentExpiredAt.Round(time.Second),
			TaskID:            "1",
			CreatedAt:         time.Time{},
			UpdatedAt:         sql.NullTime{},
			DeletedAt:         sql.NullTime{},
		}

		// mock repo
		repoMock.On("CheckStockTicket", ctx, mockPayload.TicketDetailID).Return(int64(100), nil)
		repoMock.On("DecrementStockTicket", ctx, mockPayload.TicketDetailID).Return(nil)
		repoMock.On("UpsertBooking", ctx, &mockBooking).Return(mockBooking.ID.String(), nil)
		repoMock.On("InquiryTicketAmount", ctx, mockPayload.TicketDetailID, mockPayload.TotalTickets).Return(float64(100), nil)
		repoMock.On("SetTaskScheduler", ctx, expiredAt.Round(time.Second), jsonPayloadScheduler).Return("1", nil)
		repoMock.On("UpsertPayment", ctx, &paymentMockUpsert).Return(nil)

		// test
		err := uc.ConsumeBookTicketQueue(ctx, &mockPayload)

		// assert
		assert.NoError(t, err)
	})
}

func TestShowBooking(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// mock data
		mockBooking := entity.Booking{
			ID:             uuid.UUID{},
			UserID:         1,
			TicketDetailID: 1,
			TotalTickets:   1,
			FullName:       "test",
			PersonalID:     "123",
			BookingDate:    dateTimeNow,
		}

		paymentMock := entity.Payment{
			ID:                1,
			BookingID:         uuid.UUID{},
			Amount:            1000,
			Currency:          "USD",
			Status:            "pending",
			PaymentMethod:     "Paypal",
			PaymentDate:       dateTimeNow.Round(time.Second),
			PaymentExpiration: time.Time{},
			TaskID:            "1",
			CreatedAt:         time.Time{},
			UpdatedAt:         sql.NullTime{},
			DeletedAt:         sql.NullTime{},
		}

		expectedResponse := response.BookedTicket{
			ID:            mockBooking.ID.String(),
			FullName:      mockBooking.FullName,
			PersonalID:    mockBooking.PersonalID,
			BookingDate:   mockBooking.BookingDate.Format("2006-01-02 15:04:05"),
			PaymentExpiry: paymentMock.PaymentExpiration.Format("2006-01-02 15:04:05"),
			TotalAmount:   1000,
			PaymentMethod: paymentMock.PaymentMethod,
			Status:        paymentMock.Status,
		}

		// mock repo
		repoMock.On("FindBookingByUserID", ctx, mockBooking.UserID).Return(mockBooking, nil)
		repoMock.On("FindPaymentByBookingID", ctx, mockBooking.ID.String()).Return(paymentMock, nil)

		// test
		response, err := uc.ShowBookings(ctx, mockBooking.UserID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestSetPaymentExpired(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// mock data
		payloadMock := request.PaymentExpiration{
			BookingID:      "00000000-0000-0000-0000-000000000000",
			TicketDetailID: 1,
			TotalTickets:   1,
		}

		paymentMock := entity.Payment{
			ID:                1,
			BookingID:         uuid.UUID{},
			Amount:            1000,
			Currency:          "USD",
			Status:            "pending",
			PaymentMethod:     "Paypal",
			PaymentDate:       dateTimeNow.Round(time.Second),
			PaymentExpiration: time.Time{},
			TaskID:            "1",
			CreatedAt:         time.Time{},
			UpdatedAt:         sql.NullTime{},
			DeletedAt:         sql.NullTime{},
		}

		paymentMockUpsert := entity.Payment{
			ID:                1,
			BookingID:         uuid.UUID{},
			Amount:            1000,
			Currency:          "USD",
			Status:            "expired",
			PaymentMethod:     "Paypal",
			PaymentDate:       dateTimeNow.Round(time.Second),
			PaymentExpiration: time.Time{},
			TaskID:            "1",
			CreatedAt:         time.Time{},
			UpdatedAt:         sql.NullTime{},
			DeletedAt:         sql.NullTime{},
		}

		// mock repo
		repoMock.On("FindPaymentByBookingID", ctx, payloadMock.BookingID).Return(paymentMock, nil)
		repoMock.On("UpsertPayment", ctx, &paymentMockUpsert).Return(nil)
		repoMock.On("DeleteTaskScheduler", ctx, paymentMock.TaskID).Return(nil)
		repoMock.On("IncrementStockTicket", ctx, int64(1)).Return(nil)

		// test
		err := uc.SetPaymentExpired(ctx, &payloadMock)

		// assert
		assert.NoError(t, err)
	})
}
