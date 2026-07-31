package middleware

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/DiegoPorta20/go-matriz-api/internal/application/auth"
	"github.com/DiegoPorta20/go-matriz-api/internal/application/factorization"
	"github.com/DiegoPorta20/go-matriz-api/internal/domain/matrix"
	"github.com/DiegoPorta20/go-matriz-api/internal/presentation/http/dto"
)

const unexpectedErrorMessage = "Unexpected error while processing the request"

type mappedError struct {
	status  int
	message string
	details []string
}

func NewErrorMapper(log *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		mapped := classify(err)

		if mapped.status == fiber.StatusInternalServerError {
			log.ErrorContext(c.Context(), "unhandled request failure",
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.String("error", err.Error()),
			)
		}

		return c.Status(mapped.status).JSON(dto.NewErrorResponse(mapped.message, mapped.details...))
	}
}

func classify(err error) mappedError {
	var validation *matrix.ValidationError
	if errors.As(err, &validation) {
		return mappedError{
			status:  fiber.StatusUnprocessableEntity,
			message: "Invalid matrix",
			details: []string{validation.Reason},
		}
	}

	if errors.Is(err, auth.ErrInvalidCredentials) {
		return mappedError{status: fiber.StatusUnauthorized, message: "Invalid username or password"}
	}

	if errors.Is(err, factorization.ErrStatisticsUnavailable) {
		return mappedError{status: fiber.StatusBadGateway, message: "Statistics service is unavailable"}
	}

	var fiberError *fiber.Error
	if errors.As(err, &fiberError) {
		return mappedError{status: fiberError.Code, message: fiberError.Message}
	}

	return mappedError{status: fiber.StatusInternalServerError, message: unexpectedErrorMessage}
}
