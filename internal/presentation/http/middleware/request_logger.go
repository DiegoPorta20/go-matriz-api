package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func NewRequestLogger(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		startedAt := time.Now()

		err := c.Next()

		log.LogAttrs(c.Context(), slog.LevelInfo, "http request",
			slog.String("requestId", requestid.FromContext(c)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", resolveStatus(c, err)),
			slog.Int64("durationUs", time.Since(startedAt).Microseconds()),
		)

		return err
	}
}

func resolveStatus(c fiber.Ctx, err error) int {
	if err == nil {
		return c.Response().StatusCode()
	}
	return classify(err).status
}
