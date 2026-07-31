package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/DiegoPorta20/go-matriz-api/docs"
	"github.com/DiegoPorta20/go-matriz-api/internal/infrastructure/token"
	"github.com/DiegoPorta20/go-matriz-api/internal/presentation/http/controllers"
	"github.com/DiegoPorta20/go-matriz-api/internal/presentation/http/middleware"
)

type Dependencies struct {
	Health        controllers.HealthController
	Auth          *controllers.AuthController
	Factorization *controllers.FactorizationController
	TokenService  *token.Service
}

func Register(app *fiber.App, deps Dependencies) {
	app.Get("/health", deps.Health.Check)

	app.Get("/swagger/*", adaptor.HTTPHandler(httpSwagger.WrapHandler))

	requireAccessToken := middleware.NewJwtMiddleware(deps.TokenService)

	api := app.Group("/api/v1")
	api.Post("/auth/login", deps.Auth.Login)

	api.Post("/factorization", requireAccessToken, deps.Factorization.Factorize)
}
