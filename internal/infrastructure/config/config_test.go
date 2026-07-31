package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DiegoPorta20/go-matriz-api/internal/infrastructure/config"
)

const validSecret = "a-configuration-secret-of-32-chars!"

func setValidEnvironment(t *testing.T) {
	t.Helper()

	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("NODE_SERVICE_URL", "http://node-api:3000")
	t.Setenv("AUTH_USERNAME", "demo")
	t.Setenv("AUTH_PASSWORD", "a-demo-password")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:4200")

	// Las opcionales se vacian para que los tests de valores por defecto midan el default y no lo
	// que traiga el entorno de quien ejecuta la suite: un runner de CI o una terminal donde se haya
	// cargado el .env del proyecto. Load() trata la cadena vacia igual que la variable ausente.
	for _, key := range []string{
		"GO_API_PORT",
		"NODE_SERVICE_TIMEOUT_MS",
		"NODE_SERVICE_AUTH",
		"JWT_EXPIRATION_MINUTES",
		"MAX_MATRIX_DIMENSION",
		"LOG_LEVEL",
		"AWS_REGION",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadReadsACompleteEnvironment(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("GO_API_PORT", "9090")
	t.Setenv("NODE_SERVICE_TIMEOUT_MS", "1500")
	t.Setenv("JWT_EXPIRATION_MINUTES", "15")
	t.Setenv("MAX_MATRIX_DIMENSION", "10")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:4200, https://app.example.com")

	configuration, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, 9090, configuration.Port)
	assert.Equal(t, 1500*time.Millisecond, configuration.NodeServiceTimeout)
	assert.Equal(t, 15*time.Minute, configuration.JWTExpiration)
	assert.Equal(t, 10, configuration.MaxMatrixDimension)
	assert.Equal(t, "debug", configuration.LogLevel)
	assert.Equal(t,
		[]string{"http://localhost:4200", "https://app.example.com"},
		configuration.CORSAllowedOrigins,
		"origins are trimmed",
	)
}

func TestLoadAppliesDefaultsForOptionalValues(t *testing.T) {
	setValidEnvironment(t)

	configuration, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, 8080, configuration.Port)
	assert.Equal(t, 5*time.Second, configuration.NodeServiceTimeout)
	assert.Equal(t, time.Hour, configuration.JWTExpiration)
	assert.Equal(t, 50, configuration.MaxMatrixDimension)
	assert.Equal(t, "info", configuration.LogLevel)
}

func TestLoadReportsAllMissingVariablesTogether(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("NODE_SERVICE_URL", "")
	t.Setenv("AUTH_USERNAME", "")
	t.Setenv("AUTH_PASSWORD", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	_, err := config.Load()

	require.Error(t, err)
	for _, key := range []string{
		"JWT_SECRET", "NODE_SERVICE_URL", "AUTH_USERNAME", "AUTH_PASSWORD", "CORS_ALLOWED_ORIGINS",
	} {
		assert.Contains(t, err.Error(), key)
	}
}

func TestLoadRejectsUnusableValues(t *testing.T) {
	cases := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{"secret too short", "JWT_SECRET", "too-short", "JWT_SECRET"},
		{"wildcard origin", "CORS_ALLOWED_ORIGINS", "*", "CORS_ALLOWED_ORIGINS"},
		{"wildcard among origins", "CORS_ALLOWED_ORIGINS", "http://localhost:4200,*", "CORS_ALLOWED_ORIGINS"},
		{"origins are only commas", "CORS_ALLOWED_ORIGINS", ",,", "CORS_ALLOWED_ORIGINS"},
		{"port is not a number", "GO_API_PORT", "eighty", "GO_API_PORT"},
		{"port is zero", "GO_API_PORT", "0", "GO_API_PORT"},
		{"negative timeout", "NODE_SERVICE_TIMEOUT_MS", "-100", "NODE_SERVICE_TIMEOUT_MS"},
		{"dimension is zero", "MAX_MATRIX_DIMENSION", "0", "MAX_MATRIX_DIMENSION"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv(testCase.key, testCase.value)

			_, err := config.Load()

			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expected)
		})
	}
}

func TestLoadDefaultsToCallingNodeApiWithoutSigning(t *testing.T) {
	setValidEnvironment(t)

	configuration, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, config.NodeServiceAuthNone, configuration.NodeServiceAuth)
}

func TestLoadAcceptsIamSigningWhenTheRegionIsKnown(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("NODE_SERVICE_AUTH", "IAM")
	t.Setenv("AWS_REGION", "eu-west-1")

	configuration, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, config.NodeServiceAuthIAM, configuration.NodeServiceAuth, "el valor se normaliza")
	assert.Equal(t, "eu-west-1", configuration.AWSRegion)
}

func TestLoadRefusesIamSigningWithoutARegion(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("NODE_SERVICE_AUTH", "iam")
	t.Setenv("AWS_REGION", "")

	_, err := config.Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "AWS_REGION")
}

func TestLoadRejectsAnUnknownAuthMode(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("NODE_SERVICE_AUTH", "sigv4")

	_, err := config.Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "NODE_SERVICE_AUTH")
}

func TestLoadTrimsSurroundingWhitespace(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("AUTH_USERNAME", "  demo  ")

	configuration, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "demo", configuration.AuthUsername)
}
