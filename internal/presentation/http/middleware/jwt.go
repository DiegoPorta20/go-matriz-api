package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/token"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

func NewJwtMiddleware(validator *token.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get(authorizationHeader)
		if !strings.HasPrefix(header, bearerPrefix) {
			return fiber.NewError(fiber.StatusUnauthorized,
				"Authorization header must use the Bearer scheme")
		}

		accessToken := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
		if err := validator.Validate(accessToken); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized,
				"Access token is invalid or has expired")
		}

		c.SetContext(token.WithAccessToken(c.Context(), accessToken))

		return c.Next()
	}
}
