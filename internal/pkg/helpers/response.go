package helpers

import (
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/log"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Result common output
type Result struct {
	Data     interface{}
	MetaData interface{}
	Error    error
	Count    int64
}

type response struct {
	Meta interface{} `json:"meta"`
	Data interface{} `json:"data"`
}

type MetaInternaAudit struct {
	Method        string    `json:"method"`
	Url           string    `json:"url"`
	Code          string    `json:"code"`
	ContentLength int64     `json:"content_length"`
	Date          time.Time `json:"date"`
	Ip            string    `json:"ip"`
}

type MetaResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MetaPaginationResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
	TotalPage int    `json:"total_page"`
	TotalData int    `json:"total_data"`
}

func getErrorStatusCode(err error) int {
	errString, ok := err.(*errors.ErrorString)
	if ok {
		return errString.Code()
	}

	// default http status code
	return http.StatusInternalServerError
}

func RespSuccess(c *fiber.Ctx, log log.Logger, data interface{}, message string) error {
	ip := c.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.IP()
	}
	meta := MetaInternaAudit{
		Date:          time.Now(),
		Url:           c.Path(),
		Method:        c.Method(),
		Code:          fmt.Sprintf("%v", fiber.StatusOK),
		ContentLength: int64(c.Request().Header.ContentLength()),
		Ip:            ip,
	}

	log.Info(c.Context(), "audit-log", fmt.Sprintf("%+v", meta))

	return c.JSON(response{
		Meta: MetaResponse{
			Code:    fiber.StatusOK,
			Message: message,
		},
		Data: data,
	})
}

func RespError(c *fiber.Ctx, log log.Logger, err error) error {
	ip := c.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.IP()
	}
	meta := MetaInternaAudit{
		Date:          time.Now(),
		Url:           c.Path(),
		Method:        c.Method(),
		Code:          fmt.Sprintf("%v", getErrorStatusCode(err)),
		Ip:            ip,
		ContentLength: int64(c.Request().Header.ContentLength()),
	}

	log.Info(c.Context(), "audit-log", fmt.Sprintf("%+v", meta))

	return c.Status(getErrorStatusCode(err)).JSON(response{
		Meta: MetaResponse{
			Code:    getErrorStatusCode(err),
			Message: err.Error(),
		},
		Data: nil,
	})
}

func RespPagination(c *fiber.Ctx, log log.Logger, data interface{}, metadata MetaPaginationResponse, message string) error {
	ip := c.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.IP()
	}
	meta := MetaInternaAudit{
		Date:          time.Now(),
		Url:           c.Path(),
		Method:        c.Method(),
		Code:          fmt.Sprintf("%v", fiber.StatusOK),
		ContentLength: int64(c.Request().Header.ContentLength()),
		Ip:            ip,
	}

	log.Info(c.Context(), "audit-log", fmt.Sprintf("%+v", meta))

	return c.JSON(response{
		Meta: MetaPaginationResponse{
			Code:      fiber.StatusOK,
			Message:   message,
			Page:      metadata.Page,
			Size:      metadata.Size,
			TotalPage: metadata.TotalPage,
			TotalData: metadata.TotalData,
		},
		Data: data,
	})
}

func RespErrorWithData(c *fiber.Ctx, log log.Logger, data interface{}, err error) error {
	ip := c.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.IP()
	}
	meta := MetaInternaAudit{
		Date:          time.Now(),
		Url:           c.Path(),
		Method:        c.Method(),
		Code:          fmt.Sprintf("%v", getErrorStatusCode(err)),
		Ip:            ip,
		ContentLength: int64(c.Request().Header.ContentLength()),
	}

	log.Info(c.Context(), "audit-log", fmt.Sprintf("%+v", meta))

	return c.Status(getErrorStatusCode(err)).JSON(response{
		Meta: MetaResponse{
			Code:    getErrorStatusCode(err),
			Message: err.Error(),
		},
		Data: data,
	})
}

func RespCustomError(c *fiber.Ctx, log log.Logger, err error) error {
	ip := c.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.IP()
	}
	meta := MetaInternaAudit{
		Date:          time.Now(),
		Url:           c.Path(),
		Method:        c.Method(),
		Code:          fmt.Sprintf("%v", getErrorStatusCode(err)),
		Ip:            ip,
		ContentLength: int64(c.Request().Header.ContentLength()),
	}

	log.Info(c.Context(), "audit-log", fmt.Sprintf("%+v", meta))

	errString, ok := err.(*errors.ErrorString)
	metaErrorCode := 500
	if ok {
		if errString.HttpCode() != 0 {
			metaErrorCode = errString.HttpCode()
		} else {
			metaErrorCode = errString.Code()
		}
	}
	return c.Status(metaErrorCode).JSON(response{
		Meta: MetaResponse{
			Code:    getErrorStatusCode(err),
			Message: err.Error(),
		},
		Data: nil,
	})
}
