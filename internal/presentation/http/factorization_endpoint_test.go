package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/detecta/reto-tecnico/go-api/internal/application/factorization"
)

func TestFactorizationRequiresAValidToken(t *testing.T) {
	cases := []struct {
		name   string
		header string
	}{
		{"no header", ""},
		{"wrong scheme", "Basic dXNlcjpwYXNz"},
		{"bearer without token", "Bearer "},
		{"garbage token", "Bearer not-a-real-token"},
		{"token signed with another secret", "Bearer " + foreignToken(t)},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			response := post(t, newTestApp(workingProvider()),
				"/api/v1/factorization", `{"matrix":[[1,2],[3,4]]}`, testCase.header)

			require.Equal(t, http.StatusUnauthorized, response.StatusCode)
			assertErrorEnvelope(t, bodyOf(t, response))
		})
	}
}

func TestFactorizationReturnsTheConsolidatedResponse(t *testing.T) {
	app := newTestApp(workingProvider())

	response := post(t, app, "/api/v1/factorization",
		`{"matrix":[[1,2],[3,4]]}`, bearer(t, app))

	require.Equal(t, http.StatusOK, response.StatusCode)

	var decoded struct {
		Success bool `json:"success"`
		Data    struct {
			Original   [][]float64 `json:"original"`
			Q          [][]float64 `json:"q"`
			R          [][]float64 `json:"r"`
			Statistics struct {
				Q struct {
					Maximum    float64 `json:"max"`
					IsDiagonal bool    `json:"isDiagonal"`
				} `json:"q"`
				R struct {
					Sum float64 `json:"sum"`
				} `json:"r"`
			} `json:"statistics"`
		} `json:"data"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	}
	require.NoError(t, json.Unmarshal([]byte(bodyOf(t, response)), &decoded))

	assert.True(t, decoded.Success)
	assert.Equal(t, [][]float64{{1, 2}, {3, 4}}, decoded.Data.Original)
	assert.Len(t, decoded.Data.Q, 2, "Q is m×m")
	assert.Len(t, decoded.Data.R, 2, "R is m×n")
	assert.NotEmpty(t, decoded.Message)
	assert.NotEmpty(t, decoded.Timestamp)

	assert.InDelta(t, 42.0, decoded.Data.Statistics.Q.Maximum, 0)
	assert.True(t, decoded.Data.Statistics.Q.IsDiagonal)
	assert.InDelta(t, 7.5, decoded.Data.Statistics.R.Sum, 0)
}

func TestFactorizationSeparatesMalformedRequestsFromInvalidMatrices(t *testing.T) {
	cases := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{"body is not json", `[[1,2]]bad`, http.StatusBadRequest},
		{"matrix field missing", `{}`, http.StatusBadRequest},
		{"matrix is null", `{"matrix":null}`, http.StatusBadRequest},
		{"matrix values are not numbers", `{"matrix":[["a","b"]]}`, http.StatusBadRequest},
		{"matrix is empty", `{"matrix":[]}`, http.StatusUnprocessableEntity},
		{"row is empty", `{"matrix":[[]]}`, http.StatusUnprocessableEntity},
		{"rows of different length", `{"matrix":[[1,2],[3]]}`, http.StatusUnprocessableEntity},
		{"more columns than rows", `{"matrix":[[1,2,3]]}`, http.StatusUnprocessableEntity},
		{"larger than the configured limit", `{"matrix":[[1,2,3,4],[1,2,3,4],[1,2,3,4],[1,2,3,4]]}`,
			http.StatusUnprocessableEntity},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			app := newTestApp(workingProvider())

			response := post(t, app, "/api/v1/factorization", testCase.body, bearer(t, app))

			require.Equal(t, testCase.expectedStatus, response.StatusCode)
			assertErrorEnvelope(t, bodyOf(t, response))
		})
	}
}

func TestFactorizationReportsAnUnavailableStatisticsServiceAsBadGateway(t *testing.T) {
	app := newTestApp(stubStatisticsProvider{err: factorization.ErrStatisticsUnavailable})

	response := post(t, app, "/api/v1/factorization", `{"matrix":[[1,2],[3,4]]}`, bearer(t, app))

	require.Equal(t, http.StatusBadGateway, response.StatusCode)

	body := bodyOf(t, response)
	assertErrorEnvelope(t, body)
	assert.NotContains(t, body, "node-api", "the upstream must not be named in a client-facing error")
}
