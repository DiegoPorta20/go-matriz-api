package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const signingAlgorithm = "HS256"

var ErrInvalidToken = errors.New("invalid access token")

type Service struct {
	secret     []byte
	expiration time.Duration
}

func NewService(secret string, expiration time.Duration) *Service {
	return &Service{secret: []byte(secret), expiration: expiration}
}

func (s *Service) Issue(subject string) (string, time.Duration, error) {
	issuedAt := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(s.expiration)),
	}

	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}

	return raw, s.expiration, nil
}

func (s *Service) Validate(raw string) error {
	parsed, err := jwt.ParseWithClaims(
		raw,
		&jwt.RegisteredClaims{},
		func(*jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{signingAlgorithm}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return fmt.Errorf("%w: subject claim is missing", ErrInvalidToken)
	}

	return nil
}
