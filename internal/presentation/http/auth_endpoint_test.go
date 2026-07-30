package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginReturnsAnAccessToken(t *testing.T) {
	app := newTestApp(workingProvider())

	response := post(t, app, "/api/v1/auth/login",
		`{"username":"`+testUsername+`","password":"`+testPassword+`"}`, "")

	require.Equal(t, http.StatusOK, response.StatusCode)

	var decoded struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken string `json:"accessToken"`
			TokenType   string `json:"tokenType"`
			ExpiresIn   int    `json:"expiresIn"`
		} `json:"data"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	}
	require.NoError(t, json.Unmarshal([]byte(bodyOf(t, response)), &decoded))

	assert.True(t, decoded.Success)
	assert.NotEmpty(t, decoded.Data.AccessToken)
	assert.Equal(t, "Bearer", decoded.Data.TokenType)
	assert.Positive(t, decoded.Data.ExpiresIn)
	assert.NotEmpty(t, decoded.Message)
	assert.NotEmpty(t, decoded.Timestamp)
}

func TestLoginRejectsBadInput(t *testing.T) {
	cases := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{"wrong password", `{"username":"demo","password":"wrong"}`, http.StatusUnauthorized},
		{"unknown user", `{"username":"intruder","password":"whatever"}`, http.StatusUnauthorized},
		{"no credentials", `{}`, http.StatusUnauthorized},
		{"password is a prefix of the real one", `{"username":"demo","password":"an-integration"}`,
			http.StatusUnauthorized},
		{"body is not json", `not json at all`, http.StatusBadRequest},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			response := post(t, newTestApp(workingProvider()), "/api/v1/auth/login", testCase.body, "")

			require.Equal(t, testCase.expectedStatus, response.StatusCode)

			body := bodyOf(t, response)
			assertErrorEnvelope(t, body)
			assert.NotContains(t, body, testPassword,
				"a failed login must not echo the configured password")
		})
	}
}
