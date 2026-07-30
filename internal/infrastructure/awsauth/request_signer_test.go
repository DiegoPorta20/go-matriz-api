package awsauth_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/awsauth"
)

const testRegion = "eu-west-1"

func withStaticCredentials(t *testing.T) {
	t.Helper()

	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
}

func newRequest(t *testing.T, payload []byte) *http.Request {
	t.Helper()

	request, err := http.NewRequest(
		http.MethodPost,
		"https://abc123.lambda-url.eu-west-1.on.aws/api/v1/statistics",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")

	return request
}

func TestSignWritesASigV4AuthorizationHeader(t *testing.T) {
	withStaticCredentials(t)

	signer, err := awsauth.NewRequestSigner(context.Background(), testRegion)
	require.NoError(t, err)

	payload := []byte(`{"q":[[1,0],[0,1]],"r":[[5,6],[0,7]]}`)
	request := newRequest(t, payload)

	require.NoError(t, signer.Sign(context.Background(), request, payload))

	authorization := request.Header.Get("Authorization")
	assert.Contains(t, authorization, "AWS4-HMAC-SHA256")
	assert.Contains(t, authorization, "Credential=AKIAIOSFODNN7EXAMPLE/")
	assert.Contains(t, authorization, "/eu-west-1/lambda/aws4_request")
	assert.Contains(t, authorization, "SignedHeaders=")
	assert.Contains(t, authorization, "Signature=")

	assert.NotEmpty(t, request.Header.Get("X-Amz-Date"))
}

func TestSignCoversThePayload(t *testing.T) {
	withStaticCredentials(t)

	signer, err := awsauth.NewRequestSigner(context.Background(), testRegion)
	require.NoError(t, err)

	primerPayload := []byte(`{"q":[[1]],"r":[[1]]}`)
	segundoPayload := []byte(`{"q":[[9]],"r":[[9]]}`)

	primera := newRequest(t, primerPayload)
	segunda := newRequest(t, segundoPayload)

	require.NoError(t, signer.Sign(context.Background(), primera, primerPayload))
	require.NoError(t, signer.Sign(context.Background(), segunda, segundoPayload))

	assert.NotEqual(t,
		primera.Header.Get("Authorization"),
		segunda.Header.Get("Authorization"),
		"dos cuerpos distintos no pueden compartir firma")
}

func TestSignCoversTheAccessTokenHeader(t *testing.T) {
	withStaticCredentials(t)

	signer, err := awsauth.NewRequestSigner(context.Background(), testRegion)
	require.NoError(t, err)

	payload := []byte(`{"q":[[1]],"r":[[1]]}`)

	conToken := newRequest(t, payload)
	conToken.Header.Set("X-Access-Token", "el-token-del-usuario")

	conOtroToken := newRequest(t, payload)
	conOtroToken.Header.Set("X-Access-Token", "otro-token")

	require.NoError(t, signer.Sign(context.Background(), conToken, payload))
	require.NoError(t, signer.Sign(context.Background(), conOtroToken, payload))

	assert.Contains(t, conToken.Header.Get("Authorization"), "x-access-token")
	assert.NotEqual(t,
		conToken.Header.Get("Authorization"),
		conOtroToken.Header.Get("Authorization"),
		"cambiar el token tiene que invalidar la firma")
}

func TestSignPreservesTheBodyForSending(t *testing.T) {
	withStaticCredentials(t)

	signer, err := awsauth.NewRequestSigner(context.Background(), testRegion)
	require.NoError(t, err)

	payload := []byte(`{"q":[[1]],"r":[[1]]}`)
	request := newRequest(t, payload)

	require.NoError(t, signer.Sign(context.Background(), request, payload))

	enviado := make([]byte, len(payload))
	read, err := request.Body.Read(enviado)
	require.NoError(t, err)
	assert.Equal(t, len(payload), read)
	assert.Equal(t, payload, enviado)
}

func TestNewRequestSignerUsesTheRegionItIsGiven(t *testing.T) {
	withStaticCredentials(t)

	signer, err := awsauth.NewRequestSigner(context.Background(), "us-east-2")
	require.NoError(t, err)

	payload := []byte(`{}`)
	request := newRequest(t, payload)
	require.NoError(t, signer.Sign(context.Background(), request, payload))

	assert.Contains(t, request.Header.Get("Authorization"), "/us-east-2/lambda/aws4_request")
}
