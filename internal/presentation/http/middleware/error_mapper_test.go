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

	"github.com/detecta/reto-tecnico/go-api/internal/application/auth"
	"github.com/detecta/reto-tecnico/go-api/internal/application/factorization"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
	"github.com/detecta/reto-tecnico/go-api/internal/presentation/http/middleware"
)

type errorEnvelope struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Errors    []string `json:"errors"`
	Timestamp string   `json:"timestamp"`
}

func TestErrorMapperTranslatesEachKindOfFailure(t *testing.T) {
	cases := []struct {
		name            string
		raised          error
		expectedStatus  int
		expectedMessage string
		expectedDetails []string
	}{
		{
			name:            "invalid matrix",
			raised:          &matrix.ValidationError{Reason: "Matrix must be rectangular"},
			expectedStatus:  http.StatusUnprocessableEntity,
			expectedMessage: "Invalid matrix",
			expectedDetails: []string{"Matrix must be rectangular"},
		},
		{
			name:            "invalid credentials",
			raised:          auth.ErrInvalidCredentials,
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Invalid username or password",
			expectedDetails: []string{},
		},
		{
			name:            "statistics service down",
			raised:          factorization.ErrStatisticsUnavailable,
			expectedStatus:  http.StatusBadGateway,
			expectedMessage: "Statistics service is unavailable",
			expectedDetails: []string{},
		},
		{
			name:            "transport error from a controller",
			raised:          fiber.NewError(http.StatusBadRequest, "Field matrix is required"),
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Field matrix is required",
			expectedDetails: []string{},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			response, body, _ := requestFailingWith(t, testCase.raised)

			require.Equal(t, testCase.expectedStatus, response.StatusCode)
			assert.False(t, body.Success)
			assert.Equal(t, testCase.expectedMessage, body.Message)
			assert.Equal(t, testCase.expectedDetails, body.Errors)
			assert.NotEmpty(t, body.Timestamp)
		})
	}
}

func TestErrorMapperHidesUnrecognisedErrorsBehindA500(t *testing.T) {
	raised := errors.New("dial tcp 10.0.0.5:5432: connect: connection refused")

	response, body, logged := requestFailingWith(t, raised)

	require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Equal(t, "Unexpected error while processing the request", body.Message)
	assert.Empty(t, body.Errors)
	assert.NotContains(t, body.Message, "10.0.0.5")
	assert.NotContains(t, body.Message, "connection refused")

	assert.Contains(t, logged, "connection refused")
	assert.Contains(t, logged, "unhandled request failure")
}

func TestErrorMapperDoesNotLogExpectedFailures(t *testing.T) {
	_, _, logged := requestFailingWith(t, &matrix.ValidationError{Reason: "Matrix must be rectangular"})

	assert.NotContains(t, logged, "unhandled request failure")
}

func requestFailingWith(t *testing.T, raised error) (*http.Response, errorEnvelope, string) {
	t.Helper()

	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))

	app := fiber.New(fiber.Config{ErrorHandler: middleware.NewErrorMapper(log)})
	app.Post("/resource", func(fiber.Ctx) error { return raised })

	request, err := http.NewRequest(http.MethodPost, "/resource", nil)
	require.NoError(t, err)

	response, err := app.Test(request)
	require.NoError(t, err)
	defer response.Body.Close()

	var body errorEnvelope
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))

	return response, body, logs.String()
}
