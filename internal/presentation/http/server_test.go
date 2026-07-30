package http_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/detecta/reto-tecnico/go-api/internal/application/auth"
	"github.com/detecta/reto-tecnico/go-api/internal/application/factorization"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/statistics"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/config"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/logger"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/qr"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/token"
	presentation "github.com/detecta/reto-tecnico/go-api/internal/presentation/http"
	"github.com/detecta/reto-tecnico/go-api/internal/presentation/http/controllers"
	"github.com/detecta/reto-tecnico/go-api/internal/presentation/http/routes"
)

const (
	testSecret       = "an-integration-test-secret-32-chars"
	testUsername     = "demo"
	testPassword     = "an-integration-test-password"
	testMaxDimension = 3
)

type stubStatisticsProvider struct {
	report statistics.Report
	err    error
}

func (s stubStatisticsProvider) Calculate(
	_ context.Context,
	_, _ matrix.Matrix,
) (statistics.Report, error) {
	return s.report, s.err
}

func TestHealthIsPublic(t *testing.T) {
	response := get(t, newTestApp(workingProvider()), "/health", "")

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.JSONEq(t, `{"status":"ok"}`, bodyOf(t, response))
}

func TestSwaggerUIIsServed(t *testing.T) {
	app := newTestApp(workingProvider())

	assert.Equal(t, http.StatusOK, get(t, app, "/swagger/index.html", "").StatusCode)
	assert.Equal(t, http.StatusOK, get(t, app, "/swagger/doc.json", "").StatusCode)
}

func TestUnknownRouteAnswersWithTheErrorEnvelope(t *testing.T) {
	response := get(t, newTestApp(workingProvider()), "/api/v1/does-not-exist", "")

	require.Equal(t, http.StatusNotFound, response.StatusCode)
	assertErrorEnvelope(t, bodyOf(t, response))
}

func TestSecurityHeadersArePresent(t *testing.T) {
	response := get(t, newTestApp(workingProvider()), "/health", "")

	assert.Equal(t, "nosniff", response.Header.Get("X-Content-Type-Options"))
	assert.NotEmpty(t, response.Header.Get("X-Frame-Options"))
	assert.NotEmpty(t, response.Header.Get("Referrer-Policy"))
}

func newTestApp(provider factorization.StatisticsProvider) *fiber.App {
	configuration := config.Config{
		Port:               8080,
		NodeServiceURL:     "http://node-api:3000",
		NodeServiceTimeout: time.Second,
		JWTSecret:          testSecret,
		JWTExpiration:      time.Hour,
		AuthUsername:       testUsername,
		AuthPassword:       testPassword,
		CORSAllowedOrigins: []string{"http://localhost:4200"},
		MaxMatrixDimension: testMaxDimension,
		LogLevel:           "error",
	}

	tokenService := token.NewService(configuration.JWTSecret, configuration.JWTExpiration)

	factorizeMatrix := factorization.NewFactorizeMatrixUseCase(
		qr.NewFactorizationService(), provider, configuration.MaxMatrixDimension)
	issueAccessToken := auth.NewIssueAccessTokenUseCase(tokenService, auth.Credentials{
		Username: configuration.AuthUsername,
		Password: configuration.AuthPassword,
	})

	return presentation.NewApp(configuration, logger.New(configuration.LogLevel), routes.Dependencies{
		Health:        controllers.NewHealthController(),
		Auth:          controllers.NewAuthController(issueAccessToken),
		Factorization: controllers.NewFactorizationController(factorizeMatrix),
		TokenService:  tokenService,
	})
}

func workingProvider() stubStatisticsProvider {
	return stubStatisticsProvider{report: statistics.Report{
		Orthogonal:      statistics.MatrixStatistics{Maximum: 42, IsDiagonal: true},
		UpperTriangular: statistics.MatrixStatistics{Sum: 7.5},
	}}
}

func bearer(t *testing.T, app *fiber.App) string {
	t.Helper()

	response := post(t, app, "/api/v1/auth/login",
		`{"username":"`+testUsername+`","password":"`+testPassword+`"}`, "")
	require.Equal(t, http.StatusOK, response.StatusCode)

	var decoded struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(bodyOf(t, response)), &decoded))

	return "Bearer " + decoded.Data.AccessToken
}

func foreignToken(t *testing.T) string {
	t.Helper()

	raw, _, err := token.NewService("a-completely-different-secret-32ch", time.Hour).Issue("demo")
	require.NoError(t, err)

	return raw
}

func post(t *testing.T, app *fiber.App, path, body, authorization string) *http.Response {
	t.Helper()

	request, err := http.NewRequest(http.MethodPost, path, strings.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	response, err := app.Test(request)
	require.NoError(t, err)

	return response
}

func get(t *testing.T, app *fiber.App, path, authorization string) *http.Response {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, path, nil)
	require.NoError(t, err)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	response, err := app.Test(request)
	require.NoError(t, err)

	return response
}

func bodyOf(t *testing.T, response *http.Response) string {
	t.Helper()

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return string(body)
}

func assertErrorEnvelope(t *testing.T, body string) {
	t.Helper()

	var decoded struct {
		Success   *bool     `json:"success"`
		Message   string    `json:"message"`
		Errors    *[]string `json:"errors"`
		Timestamp string    `json:"timestamp"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &decoded))

	require.NotNil(t, decoded.Success)
	assert.False(t, *decoded.Success)
	assert.NotEmpty(t, decoded.Message)
	assert.NotNil(t, decoded.Errors, "errors must always be present, even if empty")
	assert.NotEmpty(t, decoded.Timestamp)
	assert.NotContains(t, decoded.Message, "goroutine", "no stack traces in responses")
}
