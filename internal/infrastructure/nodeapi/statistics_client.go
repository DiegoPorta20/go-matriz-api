package nodeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/detecta/reto-tecnico/go-api/internal/application/factorization"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/statistics"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/token"
)

const statisticsPath = "/api/v1/statistics"

// Cabecera en la que viaja el token del usuario cuando la firma SigV4 ocupa
// Authorization. node-api acepta las dos.
const accessTokenHeader = "X-Access-Token"

// RequestSigner firma la peticion saliente. Solo se usa cuando node-api esta
// detras de una Function URL con autenticacion IAM; en la red interna de Docker
// no hay nada que firmar.
type RequestSigner interface {
	Sign(ctx context.Context, request *http.Request, payload []byte) error
}

type StatisticsClient struct {
	baseURL    string
	httpClient *http.Client
	signer     RequestSigner
}

// NewStatisticsClient construye el cliente sin firma, para cuando node-api es
// alcanzable directamente.
func NewStatisticsClient(baseURL string, timeout time.Duration) *StatisticsClient {
	return &StatisticsClient{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

// NewSignedStatisticsClient construye el cliente que firma cada peticion con
// SigV4, para cuando node-api solo acepta llamadas autenticadas por IAM.
func NewSignedStatisticsClient(
	baseURL string,
	timeout time.Duration,
	signer RequestSigner,
) *StatisticsClient {
	client := NewStatisticsClient(baseURL, timeout)
	client.signer = signer
	return client
}

func (c *StatisticsClient) Calculate(
	ctx context.Context,
	orthogonal, upperTriangular matrix.Matrix,
) (statistics.Report, error) {
	payload, err := json.Marshal(statisticsRequest{Q: orthogonal, R: upperTriangular})
	if err != nil {
		return statistics.Report{}, fmt.Errorf("encode statistics request: %w", err)
	}

	request, err := c.newRequest(ctx, payload)
	if err != nil {
		return statistics.Report{}, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return statistics.Report{}, fmt.Errorf("%w: %v", factorization.ErrStatisticsUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return statistics.Report{}, fmt.Errorf("%w: responded with status %d",
			factorization.ErrStatisticsUnavailable, response.StatusCode)
	}

	var decoded statisticsResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return statistics.Report{}, fmt.Errorf("%w: response body is not valid JSON: %v",
			factorization.ErrStatisticsUnavailable, err)
	}

	if !decoded.Success {
		return statistics.Report{}, fmt.Errorf("%w: reported a failure",
			factorization.ErrStatisticsUnavailable)
	}

	return decoded.toReport(), nil
}

func (c *StatisticsClient) newRequest(ctx context.Context, payload []byte) (*http.Request, error) {
	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.baseURL+statisticsPath, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build statistics request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	// El token del llamante se reenvia sin modificar. Emitir uno nuevo aqui
	// significaria que este servicio puede suplantar a cualquiera.
	//
	// Con firma va en una cabecera propia, porque SigV4 escribe en Authorization y
	// las dos no caben. Sin firma va donde siempre.
	if accessToken, ok := token.AccessTokenFrom(ctx); ok {
		if c.signer != nil {
			request.Header.Set(accessTokenHeader, accessToken)
		} else {
			request.Header.Set("Authorization", "Bearer "+accessToken)
		}
	}

	if c.signer != nil {
		// Se firma al final: SigV4 incluye las cabeceras en el calculo, asi que
		// cualquier cambio posterior invalidaria la firma.
		if err := c.signer.Sign(ctx, request, payload); err != nil {
			return nil, fmt.Errorf("sign statistics request: %w", err)
		}
	}

	return request, nil
}
