// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import entity "booking-service/internal/module/booking/models/entity"
import mock "github.com/stretchr/testify/mock"

import response "booking-service/internal/module/booking/models/response"
import time "time"

// Repositories is an autogenerated mock type for the Repositories type
type Repositories struct {
	mock.Mock
}

// CheckStockTicket provides a mock function with given fields: ctx, ticketDetailID
func (_m *Repositories) CheckStockTicket(ctx context.Context, ticketDetailID int64) (int64, error) {
	ret := _m.Called(ctx, ticketDetailID)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, int64) int64); ok {
		r0 = rf(ctx, ticketDetailID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, ticketDetailID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DecrementStockTicket provides a mock function with given fields: ctx, ticketDetailID
func (_m *Repositories) DecrementStockTicket(ctx context.Context, ticketDetailID int64) error {
	ret := _m.Called(ctx, ticketDetailID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, ticketDetailID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteTaskScheduler provides a mock function with given fields: ctx, taskID
func (_m *Repositories) DeleteTaskScheduler(ctx context.Context, taskID string) error {
	ret := _m.Called(ctx, taskID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, taskID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FindBookingByID provides a mock function with given fields: ctx, bookingID
func (_m *Repositories) FindBookingByID(ctx context.Context, bookingID string) (entity.Booking, error) {
	ret := _m.Called(ctx, bookingID)

	var r0 entity.Booking
	if rf, ok := ret.Get(0).(func(context.Context, string) entity.Booking); ok {
		r0 = rf(ctx, bookingID)
	} else {
		r0 = ret.Get(0).(entity.Booking)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, bookingID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindBookingByUserID provides a mock function with given fields: ctx, userID
func (_m *Repositories) FindBookingByUserID(ctx context.Context, userID int64) (entity.Booking, error) {
	ret := _m.Called(ctx, userID)

	var r0 entity.Booking
	if rf, ok := ret.Get(0).(func(context.Context, int64) entity.Booking); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Get(0).(entity.Booking)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindPaymentByBookingID provides a mock function with given fields: ctx, bookingID
func (_m *Repositories) FindPaymentByBookingID(ctx context.Context, bookingID string) (entity.Payment, error) {
	ret := _m.Called(ctx, bookingID)

	var r0 entity.Payment
	if rf, ok := ret.Get(0).(func(context.Context, string) entity.Payment); ok {
		r0 = rf(ctx, bookingID)
	} else {
		r0 = ret.Get(0).(entity.Payment)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, bookingID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IncrementStockTicket provides a mock function with given fields: ctx, ticketDetailID
func (_m *Repositories) IncrementStockTicket(ctx context.Context, ticketDetailID int64) error {
	ret := _m.Called(ctx, ticketDetailID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, ticketDetailID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InquiryTicketAmount provides a mock function with given fields: ctx, ticketDetailID, totalTicket
func (_m *Repositories) InquiryTicketAmount(ctx context.Context, ticketDetailID int64, totalTicket int) (float64, error) {
	ret := _m.Called(ctx, ticketDetailID, totalTicket)

	var r0 float64
	if rf, ok := ret.Get(0).(func(context.Context, int64, int) float64); ok {
		r0 = rf(ctx, ticketDetailID, totalTicket)
	} else {
		r0 = ret.Get(0).(float64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64, int) error); ok {
		r1 = rf(ctx, ticketDetailID, totalTicket)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetTaskScheduler provides a mock function with given fields: ctx, expiredAt, payload
func (_m *Repositories) SetTaskScheduler(ctx context.Context, expiredAt time.Duration, payload []byte) (string, error) {
	ret := _m.Called(ctx, expiredAt, payload)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration, []byte) string); ok {
		r0 = rf(ctx, expiredAt, payload)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, time.Duration, []byte) error); ok {
		r1 = rf(ctx, expiredAt, payload)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpsertBooking provides a mock function with given fields: ctx, booking
func (_m *Repositories) UpsertBooking(ctx context.Context, booking *entity.Booking) (string, error) {
	ret := _m.Called(ctx, booking)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *entity.Booking) string); ok {
		r0 = rf(ctx, booking)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *entity.Booking) error); ok {
		r1 = rf(ctx, booking)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpsertPayment provides a mock function with given fields: ctx, payment
func (_m *Repositories) UpsertPayment(ctx context.Context, payment *entity.Payment) error {
	ret := _m.Called(ctx, payment)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *entity.Payment) error); ok {
		r0 = rf(ctx, payment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateToken provides a mock function with given fields: ctx, token
func (_m *Repositories) ValidateToken(ctx context.Context, token string) (response.UserServiceValidate, error) {
	ret := _m.Called(ctx, token)

	var r0 response.UserServiceValidate
	if rf, ok := ret.Get(0).(func(context.Context, string) response.UserServiceValidate); ok {
		r0 = rf(ctx, token)
	} else {
		r0 = ret.Get(0).(response.UserServiceValidate)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}