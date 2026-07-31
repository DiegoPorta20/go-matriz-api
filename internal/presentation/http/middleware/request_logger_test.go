package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DiegoPorta20/go-matriz-api/internal/application/factorization"
	"github.com/DiegoPorta20/go-matriz-api/internal/domain/matrix"
	"github.com/DiegoPorta20/go-matriz-api/internal/presentation/http/middleware"
)

type loggedRequest struct {
	Message    string `json:"msg"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationUs int64  `json:"durationUs"`
}

func TestRequestLoggerRecordsTheStatusTheClientReceives(t *testing.T) {
	cases := []struct {
		name           string
		handlerError   error
		expectedStatus int
	}{
		{"successful request", nil, http.StatusOK},
		{"invalid matrix", &matrix.ValidationError{Reason: "not rectangular"}, http.StatusUnprocessableEntity},
		{"statistics unavailable", factorization.ErrStatisticsUnavailable, http.StatusBadGateway},
		{"bad request from the controller", fiber.NewError(http.StatusBadRequest, "nope"), http.StatusBadRequest},
		{"unauthorized", fiber.NewError(http.StatusUnauthorized, "nope"), http.StatusUnauthorized},
		{"unrecognised error", errors.New("something nobody mapped"), http.StatusInternalServerError},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var logged bytes.Buffer
			app := newLoggingApp(&logged, testCase.handlerError)

			request, err := http.NewRequest(http.MethodPost, "/resource", nil)
			require.NoError(t, err)
			response, err := app.Test(request)
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedStatus, response.StatusCode)

			entry := decodeLogEntry(t, logged.Bytes())
			assert.Equal(t, "http request", entry.Message)
			assert.Equal(t, http.MethodPost, entry.Method)
			assert.Equal(t, "/resource", entry.Path)
			assert.Equal(t, testCase.expectedStatus, entry.Status,
				"the logged status must match the response")
			assert.GreaterOrEqual(t, entry.DurationUs, int64(0))
		})
	}
}

func TestRequestLoggerNeverRecordsTheAuthorizationHeader(t *testing.T) {
	var logged bytes.Buffer
	app := newLoggingApp(&logged, nil)

	request, err := http.NewRequest(http.MethodPost, "/resource", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer a-secret-token-value")

	_, err = app.Test(request)
	require.NoError(t, err)

	assert.NotContains(t, logged.String(), "a-secret-token-value")
	assert.NotContains(t, logged.String(), "Authorization")
}

func newLoggingApp(output *bytes.Buffer, handlerError error) *fiber.App {
	log := slog.New(slog.NewJSONHandler(output, nil))

	app := fiber.New(fiber.Config{ErrorHandler: middleware.NewErrorMapper(log)})
	app.Use(middleware.NewRequestLogger(log))
	app.Post("/resource", func(c fiber.Ctx) error {
		if handlerError != nil {
			return handlerError
		}
		return c.SendStatus(http.StatusOK)
	})

	return app
}

func decodeLogEntry(t *testing.T, raw []byte) loggedRequest {
	t.Helper()

	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte("\n")) {
		var entry loggedRequest
		require.NoError(t, json.Unmarshal(line, &entry))
		if entry.Message == "http request" {
			return entry
		}
	}

	t.Fatalf("no request log line found in: %s", raw)
	return loggedRequest{}
}
