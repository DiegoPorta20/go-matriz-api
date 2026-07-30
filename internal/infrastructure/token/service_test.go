package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/token"
)

const secret = "a-test-secret-of-at-least-32-characters"

func TestIssuedTokensValidate(t *testing.T) {
	service := token.NewService(secret, time.Hour)

	raw, expiresIn, err := service.Issue("demo")

	require.NoError(t, err)
	assert.Equal(t, time.Hour, expiresIn)
	assert.NoError(t, service.Validate(raw))
}

func TestValidateRejectsTokensItShouldNotTrust(t *testing.T) {
	service := token.NewService(secret, time.Hour)

	otherSecret, _, err := token.NewService("a-different-secret-of-32-characters!", time.Hour).Issue("demo")
	require.NoError(t, err)

	expired, _, err := token.NewService(secret, -time.Minute).Issue("demo")
	require.NoError(t, err)

	cases := []struct {
		name string
		raw  string
	}{
		{"empty string", ""},
		{"not a token", "clearly-not-a-jwt"},
		{"signed with another secret", otherSecret},
		{"already expired", expired},
		{"signed with a different algorithm", signedWithHS512(t)},
		{"unsigned alg none", unsignedToken(t)},
		{"no expiration claim", withoutExpiration(t)},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := service.Validate(testCase.raw)

			require.ErrorIs(t, err, token.ErrInvalidToken)
		})
	}
}

func TestAccessTokenTravelsInTheContext(t *testing.T) {
	ctx := token.WithAccessToken(context.Background(), "the-raw-token")

	stored, ok := token.AccessTokenFrom(ctx)

	assert.True(t, ok)
	assert.Equal(t, "the-raw-token", stored)
}

func TestAccessTokenIsAbsentWhenNeverStored(t *testing.T) {
	_, ok := token.AccessTokenFrom(context.Background())

	assert.False(t, ok)
}

func TestEmptyAccessTokenCountsAsAbsent(t *testing.T) {
	ctx := token.WithAccessToken(context.Background(), "")

	_, ok := token.AccessTokenFrom(ctx)

	assert.False(t, ok, "an empty token must never be forwarded as if it were valid")
}

func signedWithHS512(t *testing.T) string {
	t.Helper()

	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   "demo",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(secret))
	require.NoError(t, err)

	return raw
}

func unsignedToken(t *testing.T) string {
	t.Helper()

	raw, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   "demo",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	return raw
}

func withoutExpiration(t *testing.T) string {
	t.Helper()

	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: "demo",
	}).SignedString([]byte(secret))
	require.NoError(t, err)

	return raw
}
