package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type TokenIssuer interface {
	Issue(subject string) (raw string, expiresIn time.Duration, err error)
}

type Credentials struct {
	Username string
	Password string
}

type AccessToken struct {
	Raw       string
	ExpiresIn time.Duration
}

type IssueAccessTokenUseCase struct {
	issuer   TokenIssuer
	expected Credentials
}

func NewIssueAccessTokenUseCase(issuer TokenIssuer, expected Credentials) *IssueAccessTokenUseCase {
	return &IssueAccessTokenUseCase{issuer: issuer, expected: expected}
}

func (uc *IssueAccessTokenUseCase) Execute(provided Credentials) (AccessToken, error) {
	if !uc.matches(provided) {
		return AccessToken{}, ErrInvalidCredentials
	}

	raw, expiresIn, err := uc.issuer.Issue(provided.Username)
	if err != nil {
		return AccessToken{}, fmt.Errorf("issue access token: %w", err)
	}

	return AccessToken{Raw: raw, ExpiresIn: expiresIn}, nil
}

func (uc *IssueAccessTokenUseCase) matches(provided Credentials) bool {
	usernameMatches := subtle.ConstantTimeCompare(
		[]byte(provided.Username), []byte(uc.expected.Username)) == 1
	passwordMatches := subtle.ConstantTimeCompare(
		[]byte(provided.Password), []byte(uc.expected.Password)) == 1

	return usernameMatches && passwordMatches
}
