package nodeapi_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/detecta/reto-tecnico/go-api/internal/application/factorization"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/nodeapi"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/token"
)

// Imita lo que hace SigV4: escribir en Authorization. Se firma de verdad en
// awsauth; aqui lo que se prueba es que el cliente coloca cada cosa en su sitio.
type stubSigner struct {
	err             error
	calls           int
	receivedPayload []byte
	headersAlFirmar http.Header
}

func (s *stubSigner) Sign(_ context.Context, request *http.Request, payload []byte) error {
	s.calls++
	s.receivedPayload = payload
	s.headersAlFirmar = request.Header.Clone()

	if s.err != nil {
		return s.err
	}

	request.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=test/20260730/eu-west-1/lambda/aws4_request")
	return nil
}

// Con firma, Authorization la ocupa SigV4, asi que el JWT del usuario tiene que
// viajar en su propia cabecera o se perderia.
func TestSignedClientSendsTheUserTokenInItsOwnHeader(t *testing.T) {
	var recibidas http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recibidas = r.Header.Clone()
		writeJSON(t, w, http.StatusOK, validResponse)
	}))
	defer server.Close()

	signer := &stubSigner{}
	client := nodeapi.NewSignedStatisticsClient(server.URL, clientTimeout, signer)
	ctx := token.WithAccessToken(context.Background(), "el-token-del-usuario")

	_, err := client.Calculate(ctx, orthogonal, upperTriangular)

	require.NoError(t, err)
	assert.Equal(t, "el-token-del-usuario", recibidas.Get("X-Access-Token"))
	assert.Contains(t, recibidas.Get("Authorization"), "AWS4-HMAC-SHA256")
	assert.NotContains(t, recibidas.Get("Authorization"), "Bearer")
}

// El orden importa: SigV4 incluye las cabeceras en el calculo, asi que si el
// token se añadiera despues de firmar, la firma no cuadraria.
func TestSignedClientSignsAfterSettingEveryHeader(t *testing.T) {
	server := serverReturning(t, http.StatusOK, validResponse)
	signer := &stubSigner{}
	ctx := token.WithAccessToken(context.Background(), "el-token-del-usuario")

	_, err := nodeapi.NewSignedStatisticsClient(server.URL, clientTimeout, signer).
		Calculate(ctx, orthogonal, upperTriangular)

	require.NoError(t, err)
	require.Equal(t, 1, signer.calls)
	assert.Equal(t, "el-token-del-usuario", signer.headersAlFirmar.Get("X-Access-Token"))
	assert.Equal(t, "application/json", signer.headersAlFirmar.Get("Content-Type"))
}

// SigV4 necesita el hash del cuerpo. Si se leyera de request.Body, el cuerpo
// llegaria vacio al servidor.
func TestSignedClientGivesTheSignerTheSameBodyItSends(t *testing.T) {
	var cuerpoRecibido []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cuerpoRecibido, _ = io.ReadAll(r.Body)
		writeJSON(t, w, http.StatusOK, validResponse)
	}))
	defer server.Close()

	signer := &stubSigner{}

	_, err := nodeapi.NewSignedStatisticsClient(server.URL, clientTimeout, signer).
		Calculate(context.Background(), orthogonal, upperTriangular)

	require.NoError(t, err)
	assert.NotEmpty(t, signer.receivedPayload)
	assert.Equal(t, signer.receivedPayload, cuerpoRecibido)
}

func TestSignedClientFailsWhenTheSignatureFails(t *testing.T) {
	server := serverReturning(t, http.StatusOK, validResponse)
	fallo := errors.New("no hay credenciales")

	_, err := nodeapi.NewSignedStatisticsClient(server.URL, clientTimeout, &stubSigner{err: fallo}).
		Calculate(context.Background(), orthogonal, upperTriangular)

	require.ErrorIs(t, err, fallo)
}

// El cliente sin firma es el que usa docker compose, y ahi el JWT sigue yendo
// donde siempre.
func TestUnsignedClientKeepsUsingTheBearerHeader(t *testing.T) {
	var recibidas http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recibidas = r.Header.Clone()
		writeJSON(t, w, http.StatusOK, validResponse)
	}))
	defer server.Close()

	ctx := token.WithAccessToken(context.Background(), "el-token-del-usuario")

	_, err := nodeapi.NewStatisticsClient(server.URL, clientTimeout).
		Calculate(ctx, orthogonal, upperTriangular)

	require.NoError(t, err)
	assert.Equal(t, "Bearer el-token-del-usuario", recibidas.Get("Authorization"))
	assert.Empty(t, recibidas.Get("X-Access-Token"))
}

// El cliente firmado sigue siendo un StatisticsProvider: el caso de uso no se
// entera de que hay firma.
func TestSignedClientSatisfiesThePort(t *testing.T) {
	var _ factorization.StatisticsProvider = nodeapi.NewSignedStatisticsClient(
		"http://node-api:3000", clientTimeout, &stubSigner{})
}
