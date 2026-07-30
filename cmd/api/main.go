package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/detecta/reto-tecnico/go-api/internal/application/auth"
	"github.com/detecta/reto-tecnico/go-api/internal/application/factorization"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/awsauth"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/config"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/logger"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/nodeapi"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/qr"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/token"
	presentation "github.com/detecta/reto-tecnico/go-api/internal/presentation/http"
	"github.com/detecta/reto-tecnico/go-api/internal/presentation/http/controllers"
	"github.com/detecta/reto-tecnico/go-api/internal/presentation/http/routes"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	configuration, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(configuration.LogLevel)

	app, err := buildApp(context.Background(), configuration, log)
	if err != nil {
		return err
	}

	return listenUntilSignal(app, configuration, log)
}

func buildApp(
	ctx context.Context,
	configuration config.Config,
	log *slog.Logger,
) (*fiber.App, error) {
	tokenService := token.NewService(configuration.JWTSecret, configuration.JWTExpiration)

	statisticsClient, err := buildStatisticsClient(ctx, configuration, log)
	if err != nil {
		return nil, err
	}

	factorizeMatrix := factorization.NewFactorizeMatrixUseCase(
		qr.NewFactorizationService(), statisticsClient, configuration.MaxMatrixDimension)
	issueAccessToken := auth.NewIssueAccessTokenUseCase(tokenService, auth.Credentials{
		Username: configuration.AuthUsername,
		Password: configuration.AuthPassword,
	})

	return presentation.NewApp(configuration, log, routes.Dependencies{
		Health:        controllers.NewHealthController(),
		Auth:          controllers.NewAuthController(issueAccessToken),
		Factorization: controllers.NewFactorizationController(factorizeMatrix),
		TokenService:  tokenService,
	}), nil
}

// buildStatisticsClient decide si la llamada a node-api va firmada.
//
// Es lo unico que cambia entre correr en la red interna de Docker y correr contra
// una Lambda Function URL con autenticacion IAM. Ni el caso de uso ni el dominio
// se enteran.
func buildStatisticsClient(
	ctx context.Context,
	configuration config.Config,
	log *slog.Logger,
) (factorization.StatisticsProvider, error) {
	if configuration.NodeServiceAuth != config.NodeServiceAuthIAM {
		return nodeapi.NewStatisticsClient(
			configuration.NodeServiceURL, configuration.NodeServiceTimeout), nil
	}

	signer, err := awsauth.NewRequestSigner(ctx, configuration.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("build request signer: %w", err)
	}

	log.Info("statistics requests will be signed with sigv4",
		slog.String("region", configuration.AWSRegion))

	return nodeapi.NewSignedStatisticsClient(
		configuration.NodeServiceURL, configuration.NodeServiceTimeout, signer), nil
}

func listenUntilSignal(app *fiber.App, configuration config.Config, log *slog.Logger) error {
	address := fmt.Sprintf(":%d", configuration.Port)

	serverStopped := make(chan error, 1)
	go func() {
		serverStopped <- app.Listen(address, fiber.ListenConfig{DisableStartupMessage: true})
	}()

	log.Info("service started",
		slog.Int("port", configuration.Port),
		slog.String("statisticsService", configuration.NodeServiceURL),
	)

	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverStopped:
		if err != nil {
			return fmt.Errorf("http server stopped: %w", err)
		}
		return nil
	case <-interrupted:
		log.Info("shutdown requested", slog.Duration("timeout", shutdownTimeout))
		if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		return nil
	}
}
