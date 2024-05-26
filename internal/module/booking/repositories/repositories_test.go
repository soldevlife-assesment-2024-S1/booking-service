package repositories_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	log_internal "booking-service/internal/pkg/log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"

	"booking-service/internal/module/booking/models/entity"
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/pkg/errors"
)

var (
	mock    sqlxmock.Sqlmock
	dbx     *sqlx.DB
	logMock *otelzap.Logger
)

func setup() {
	dbx, mock, _ = sqlxmock.Newx()
	logMock = log_internal.Setup()
}

func TestFindBookingByID(t *testing.T) {
	setup()
	// Create a new instance of the repository
	repo := repositories.New(dbx, logMock, nil, nil, nil, nil, nil, nil, nil)

	UUID := uuid.New()

	// Define the test case
	testCases := []struct {
		name            string
		bookingID       string
		expectedError   error
		expectedBooking entity.Booking
	}{
		{
			name:          "Booking found",
			bookingID:     UUID.String(),
			expectedError: nil,
			expectedBooking: entity.Booking{
				ID:             UUID,
				UserID:         1,
				TicketDetailID: 1,
				TotalTickets:   1,
				FullName:       "John Doe",
				PersonalID:     "1234567890",
				BookingDate:    time.Time{},
				CreatedAt:      &time.Time{},
				UpdatedAt:      sql.NullTime{},
				DeletedAt:      sql.NullTime{},
			},
		},
		{
			name:            "Booking not found",
			bookingID:       "456",
			expectedError:   nil,
			expectedBooking: entity.Booking{},
		},
		{
			name:            "Database error",
			bookingID:       "789",
			expectedError:   errors.InternalServerError("error find booking by booking id"),
			expectedBooking: entity.Booking{},
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the mock expectations
			rows := sqlxmock.NewRows([]string{
				"id", "user_id", "ticket_detail_id", "total_tickets", "full_name", "personal_id", "booking_date", "created_at", "updated_at", "deleted_at",
			}).
				AddRow(tc.expectedBooking.ID /* Add other column values here */)
			mock.ExpectQuery("SELECT * FROM bookings WHERE id = $1").
				WithArgs(tc.bookingID).
				WillReturnRows(rows)

			// Call the function under test
			booking, err := repo.FindBookingByID(context.Background(), tc.bookingID)

			// Check the result
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedBooking, booking)

			// Verify that all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
