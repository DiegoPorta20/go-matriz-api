package http

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	recovermw "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/config"
	"github.com/detecta/reto-tecnico/go-api/internal/presentation/http/middleware"
	"github.com/detecta/reto-tecnico/go-api/internal/presentation/http/routes"
)

const (
	rateLimitRequests = 60
	rateLimitWindow   = time.Minute
)

func NewApp(configuration config.Config, log *slog.Logger, deps routes.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.NewErrorMapper(log),
	})

	app.Use(recovermw.New())
	app.Use(requestid.New())
	app.Use(middleware.NewRequestLogger(log))
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: configuration.CORSAllowedOrigins,
		AllowMethods: []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodOptions},
		AllowHeaders: []string{fiber.HeaderContentType, fiber.HeaderAuthorization, fiber.HeaderAccept},
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        rateLimitRequests,
		Expiration: rateLimitWindow,
	}))

	routes.Register(app, deps)

	return app
}
