package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/detecta/reto-tecnico/go-api/internal/application/auth"
)

type stubIssuer struct {
	raw             string
	expiresIn       time.Duration
	err             error
	receivedSubject string
	calls           int
}

func (s *stubIssuer) Issue(subject string) (string, time.Duration, error) {
	s.calls++
	s.receivedSubject = subject
	return s.raw, s.expiresIn, s.err
}

func expectedCredentials() auth.Credentials {
	return auth.Credentials{Username: "demo", Password: "a-long-enough-password"}
}

func TestExecuteIssuesATokenForTheConfiguredUser(t *testing.T) {
	issuer := &stubIssuer{raw: "signed-token", expiresIn: time.Hour}
	useCase := auth.NewIssueAccessTokenUseCase(issuer, expectedCredentials())

	accessToken, err := useCase.Execute(expectedCredentials())

	require.NoError(t, err)
	assert.Equal(t, "signed-token", accessToken.Raw)
	assert.Equal(t, time.Hour, accessToken.ExpiresIn)
	assert.Equal(t, "demo", issuer.receivedSubject)
}

func TestExecuteRejectsWrongCredentialsWithoutIssuingAnything(t *testing.T) {
	cases := []struct {
		name     string
		provided auth.Credentials
	}{
		{"wrong password", auth.Credentials{Username: "demo", Password: "wrong"}},
		{"wrong username", auth.Credentials{Username: "intruder", Password: "a-long-enough-password"}},
		{"both wrong", auth.Credentials{Username: "intruder", Password: "wrong"}},
		{"empty credentials", auth.Credentials{}},
		{"password is a prefix of the real one", auth.Credentials{Username: "demo", Password: "a-long"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			issuer := &stubIssuer{raw: "must-not-be-issued"}
			useCase := auth.NewIssueAccessTokenUseCase(issuer, expectedCredentials())

			_, err := useCase.Execute(testCase.provided)

			require.ErrorIs(t, err, auth.ErrInvalidCredentials)
			assert.Zero(t, issuer.calls)
		})
	}
}

func TestExecutePropagatesASigningFailure(t *testing.T) {
	failure := errors.New("no signing key")
	useCase := auth.NewIssueAccessTokenUseCase(&stubIssuer{err: failure}, expectedCredentials())

	_, err := useCase.Execute(expectedCredentials())

	require.ErrorIs(t, err, failure)
}
