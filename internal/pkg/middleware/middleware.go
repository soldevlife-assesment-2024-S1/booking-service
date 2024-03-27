package middleware

import (
	"booking-service/internal/module/booking/repositories"
	"booking-service/internal/pkg/errors"
	"booking-service/internal/pkg/helpers"
	log "booking-service/internal/pkg/log"
	"go/token"

	"github.com/gofiber/fiber/v2"
)

type Middleware struct {
	Log  log.Logger
	Repo repositories.Repositories
}

func (m *Middleware) ValidateToken(ctx *fiber.Ctx) error {
	// get token from header
	auth := ctx.Get("Authorization")
	if auth == "" {
		m.Log.Error(ctx.Context(), "error get token from header", errors.UnauthorizedError("error get token from header"))
		return helpers.RespError(ctx, m.Log, errors.UnauthorizedError("error get token from header"))
	}

	// grab token (Bearer token) from header 7 is the length of "Bearer "
	token := auth[7:token.Pos(len(auth))]

	// check repostipories if token is valid
	resp, err := m.Repo.ValidateToken(ctx.Context(), token)
	if err != nil {
		m.Log.Error(ctx.Context(), "error validate token", err)
		return helpers.RespError(ctx, m.Log, errors.UnauthorizedError("error validate token"))
	}

	if !resp.IsValid {
		m.Log.Error(ctx.Context(), "error validate token", errors.UnauthorizedError("error validate token"))
		return helpers.RespError(ctx, m.Log, errors.UnauthorizedError("error validate token"))
	}

	ctx.Locals("user_id", resp.UserID)

	return ctx.Next()
}
