package nodeapi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DiegoPorta20/go-matriz-api/internal/application/factorization"
	"github.com/DiegoPorta20/go-matriz-api/internal/domain/matrix"
	"github.com/DiegoPorta20/go-matriz-api/internal/infrastructure/nodeapi"
	"github.com/DiegoPorta20/go-matriz-api/internal/infrastructure/token"
)

const clientTimeout = 2 * time.Second

var (
	orthogonal      = matrix.Matrix{{1, 0}, {0, 1}}
	upperTriangular = matrix.Matrix{{5, 6}, {0, 7}}
)

const validResponse = `{
  "success": true,
  "data": {
    "q": {"max": 1, "min": 0, "average": 0.5, "sum": 2, "isDiagonal": true},
    "r": {"max": 7, "min": 0, "average": 4.5, "sum": 18, "isDiagonal": false}
  }
}`

func TestCalculateMapsTheResponseIntoAReport(t *testing.T) {
	server := serverReturning(t, http.StatusOK, validResponse)
	client := nodeapi.NewStatisticsClient(server.URL, clientTimeout)

	report, err := client.Calculate(context.Background(), orthogonal, upperTriangular)

	require.NoError(t, err)
	assert.InDelta(t, 1.0, report.Orthogonal.Maximum, 0)
	assert.InDelta(t, 0.5, report.Orthogonal.Average, 0)
	assert.True(t, report.Orthogonal.IsDiagonal)
	assert.InDelta(t, 18.0, report.UpperTriangular.Sum, 0)
	assert.False(t, report.UpperTriangular.IsDiagonal)
}

func TestCalculateSendsBothMatricesAsQAndR(t *testing.T) {
	var received struct {
		Q [][]float64 `json:"q"`
		R [][]float64 `json:"r"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &received))

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/statistics", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		writeJSON(t, w, http.StatusOK, validResponse)
	}))
	defer server.Close()

	_, err := nodeapi.NewStatisticsClient(server.URL, clientTimeout).
		Calculate(context.Background(), orthogonal, upperTriangular)

	require.NoError(t, err)
	assert.Equal(t, [][]float64(orthogonal), received.Q)
	assert.Equal(t, [][]float64(upperTriangular), received.R)
}

func TestCalculateForwardsTheCallersToken(t *testing.T) {
	var receivedAuthorization string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthorization = r.Header.Get("Authorization")
		writeJSON(t, w, http.StatusOK, validResponse)
	}))
	defer server.Close()

	ctx := token.WithAccessToken(context.Background(), "the-callers-token")

	_, err := nodeapi.NewStatisticsClient(server.URL, clientTimeout).
		Calculate(ctx, orthogonal, upperTriangular)

	require.NoError(t, err)
	assert.Equal(t, "Bearer the-callers-token", receivedAuthorization)
}

func TestCalculateSendsNoAuthorizationWhenTheContextHasNoToken(t *testing.T) {
	var hadAuthorization bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hadAuthorization = r.Header["Authorization"]
		writeJSON(t, w, http.StatusOK, validResponse)
	}))
	defer server.Close()

	_, err := nodeapi.NewStatisticsClient(server.URL, clientTimeout).
		Calculate(context.Background(), orthogonal, upperTriangular)

	require.NoError(t, err)
	assert.False(t, hadAuthorization, "an empty header is worse than no header")
}

func TestCalculateReportsTheServiceAsUnavailable(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"bad request", http.StatusBadRequest, `{"success":false}`},
		{"unauthorized", http.StatusUnauthorized, `{"success":false}`},
		{"internal error", http.StatusInternalServerError, `{"success":false}`},
		{"body is not json", http.StatusOK, `<html>gateway error</html>`},
		{"success flag is false", http.StatusOK, `{"success":false,"data":{}}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			server := serverReturning(t, testCase.status, testCase.body)
			client := nodeapi.NewStatisticsClient(server.URL, clientTimeout)

			_, err := client.Calculate(context.Background(), orthogonal, upperTriangular)

			require.ErrorIs(t, err, factorization.ErrStatisticsUnavailable)
		})
	}
}

func TestCalculateReportsAnUnreachableService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	unreachableURL := server.URL
	server.Close()

	_, err := nodeapi.NewStatisticsClient(unreachableURL, clientTimeout).
		Calculate(context.Background(), orthogonal, upperTriangular)

	require.ErrorIs(t, err, factorization.ErrStatisticsUnavailable)
}

func TestCalculateGivesUpWhenTheServiceIsTooSlow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		writeJSON(t, w, http.StatusOK, validResponse)
	}))
	defer server.Close()

	client := nodeapi.NewStatisticsClient(server.URL, 30*time.Millisecond)

	_, err := client.Calculate(context.Background(), orthogonal, upperTriangular)

	require.ErrorIs(t, err, factorization.ErrStatisticsUnavailable)
}

func TestCalculateHonoursACancelledContext(t *testing.T) {
	server := serverReturning(t, http.StatusOK, validResponse)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := nodeapi.NewStatisticsClient(server.URL, clientTimeout).
		Calculate(ctx, orthogonal, upperTriangular)

	require.ErrorIs(t, err, factorization.ErrStatisticsUnavailable)
}

func TestCalculateNormalisesTheBaseURL(t *testing.T) {
	var receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		writeJSON(t, w, http.StatusOK, validResponse)
	}))
	defer server.Close()

	_, err := nodeapi.NewStatisticsClient(server.URL+"/", clientTimeout).
		Calculate(context.Background(), orthogonal, upperTriangular)

	require.NoError(t, err)
	assert.Equal(t, "/api/v1/statistics", receivedPath)
}

func serverReturning(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, status, body)
	}))
	t.Cleanup(server.Close)

	return server
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, body string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := io.WriteString(w, body)
	require.NoError(t, err)
}
