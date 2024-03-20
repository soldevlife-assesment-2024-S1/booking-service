package errors

import (
	"net/http"
)

type ErrorString struct {
	code     int
	message  string
	httpCode int
}

func (e ErrorString) Code() int {
	return e.code
}

func (e ErrorString) Error() string {
	return e.message
}

func (e ErrorString) Message() string {
	return e.message
}

func (e ErrorString) HttpCode() int {
	return e.httpCode
}

func BadRequest(msg string) error {
	return &ErrorString{
		code:    http.StatusBadRequest,
		message: msg,
	}
}

func NotFound(msg string) error {
	return &ErrorString{
		code:    http.StatusNotFound,
		message: msg,
	}
}

func Conflict(msg string) error {
	return &ErrorString{
		code:    http.StatusConflict,
		message: msg,
	}
}

func InternalServerError(msg string) error {
	return &ErrorString{
		code:    http.StatusInternalServerError,
		message: msg,
	}
}

func UnauthorizedError(msg string) error {
	return &ErrorString{
		code:    http.StatusUnauthorized,
		message: msg,
	}
}

func ForbiddenError(msg string) error {
	return &ErrorString{
		code:    http.StatusForbidden,
		message: msg,
	}
}

func CustomError(msg string, code int, codeHttp int) error {
	return &ErrorString{
		code:     code,
		message:  msg,
		httpCode: codeHttp,
	}
}

func TooManyRequest(msg string) error {
	return &ErrorString{
		code:    http.StatusTooManyRequests,
		message: msg,
	}
}

func UnprocessableEntity(msg string) error {
	return &ErrorString{
		code:    http.StatusUnprocessableEntity,
		message: msg,
	}
}
