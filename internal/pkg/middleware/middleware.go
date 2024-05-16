package middleware

import (
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/helpers"
	"fmt"
	"go/token"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type Middleware struct {
	Log  *otelzap.Logger
	Repo repositories.Repositories
}

func (m *Middleware) ValidateToken(ctx *fiber.Ctx) error {
	// get token from header
	auth := ctx.Get("Authorization")
	if auth == "" {
		m.Log.Ctx(ctx.UserContext()).Error("error get token from header")
		return helpers.RespError(ctx, m.Log, errors.UnauthorizedError("error get token from header"))
	}

	// grab token (Bearer token) from header 7 is the length of "Bearer "
	token := auth[7:token.Pos(len(auth))]

	// check repostipories if token is valid
	resp, err := m.Repo.ValidateToken(ctx.Context(), token)
	if err != nil {
		m.Log.Ctx(ctx.UserContext()).Error(fmt.Sprintf("error validate token: %v", err))
		return helpers.RespError(ctx, m.Log, errors.UnauthorizedError("error validate token"))
	}

	if !resp.IsValid {
		m.Log.Ctx(ctx.UserContext()).Error("error validate token")
		return helpers.RespError(ctx, m.Log, errors.UnauthorizedError("error validate token"))
	}

	ctx.Locals("user_id", resp.UserID)
	ctx.Locals("email_user", resp.EmailUser)

	return ctx.Next()
}

func (m *Middleware) CheckIsWeekend(ctx *fiber.Ctx) error {
	// get current day
	day := time.Now().Weekday().String()

	if day == "Saturday" || day == "Sunday" {
		return ctx.Next()
	}

	m.Log.Ctx(ctx.UserContext()).Error("error validate booking day, only can book on weekend")
	return helpers.RespError(ctx, m.Log, errors.BadRequest("error validate booking day, only can book on weekend"))
}
