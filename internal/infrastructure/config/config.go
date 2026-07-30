package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort                 = 8080
	defaultNodeTimeoutMS        = 5000
	defaultJWTExpirationMinutes = 60
	defaultMaxMatrixDimension   = 50
	defaultLogLevel             = "info"
	defaultNodeServiceAuth      = NodeServiceAuthNone

	minimumSecretLength = 32
)

type Config struct {
	Port               int
	NodeServiceURL     string
	NodeServiceTimeout time.Duration
	JWTSecret          string
	JWTExpiration      time.Duration
	AuthUsername       string
	AuthPassword       string
	CORSAllowedOrigins []string
	MaxMatrixDimension int
	LogLevel           string

	// NodeServiceAuth decide como se autentica la llamada a node-api. Con "iam"
	// la peticion se firma con SigV4, que es lo que exige una Lambda Function URL
	// con autenticacion IAM. Con "none" se llama directamente, como en la red
	// interna de Docker.
	NodeServiceAuth string
	AWSRegion       string
}

// Modos admitidos para NODE_SERVICE_AUTH.
const (
	NodeServiceAuthNone = "none"
	NodeServiceAuthIAM  = "iam"
)

func Load() (Config, error) {
	var env reader

	configuration := Config{
		Port:               env.positiveInt("GO_API_PORT", defaultPort),
		NodeServiceURL:     env.required("NODE_SERVICE_URL"),
		NodeServiceTimeout: env.milliseconds("NODE_SERVICE_TIMEOUT_MS", defaultNodeTimeoutMS),
		JWTSecret:          env.secret("JWT_SECRET"),
		JWTExpiration:      env.minutes("JWT_EXPIRATION_MINUTES", defaultJWTExpirationMinutes),
		AuthUsername:       env.required("AUTH_USERNAME"),
		AuthPassword:       env.required("AUTH_PASSWORD"),
		CORSAllowedOrigins: env.origins("CORS_ALLOWED_ORIGINS"),
		MaxMatrixDimension: env.positiveInt("MAX_MATRIX_DIMENSION", defaultMaxMatrixDimension),
		LogLevel:           env.withDefault("LOG_LEVEL", defaultLogLevel),
		NodeServiceAuth:    env.nodeServiceAuth("NODE_SERVICE_AUTH"),
		AWSRegion:          env.withDefault("AWS_REGION", ""),
	}

	// Firmar sin saber la region produce una firma que AWS rechaza, y el mensaje
	// de error no lo explica. Mejor no arrancar.
	if configuration.NodeServiceAuth == NodeServiceAuthIAM && configuration.AWSRegion == "" {
		env.reject("AWS_REGION", "is required when NODE_SERVICE_AUTH is iam")
	}

	if len(env.problems) > 0 {
		return Config{}, fmt.Errorf("invalid configuration: %s", strings.Join(env.problems, "; "))
	}

	return configuration, nil
}

type reader struct {
	problems []string
}

func (r *reader) required(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		r.reject(key, "is required")
	}
	return value
}

func (r *reader) secret(key string) string {
	value := r.required(key)
	if value != "" && len(value) < minimumSecretLength {
		r.reject(key, fmt.Sprintf("must be at least %d characters long", minimumSecretLength))
	}
	return value
}

func (r *reader) withDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (r *reader) positiveInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		r.reject(key, "must be a positive integer")
		return fallback
	}
	return value
}

func (r *reader) milliseconds(key string, fallback int) time.Duration {
	return time.Duration(r.positiveInt(key, fallback)) * time.Millisecond
}

func (r *reader) minutes(key string, fallback int) time.Duration {
	return time.Duration(r.positiveInt(key, fallback)) * time.Minute
}

func (r *reader) origins(key string) []string {
	raw := r.required(key)
	if raw == "" {
		return nil
	}

	origins := make([]string, 0)
	for _, candidate := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(candidate)
		if origin == "" {
			continue
		}
		if origin == "*" {
			r.reject(key, "must list explicit origins, not \"*\"")
			continue
		}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		r.reject(key, "must contain at least one origin")
	}
	return origins
}

// nodeServiceAuth solo admite los dos modos conocidos: un valor mal escrito
// dejaria el servicio llamando sin firma contra un endpoint que la exige, y el
// fallo apareceria en la primera peticion de un usuario y no al arrancar.
func (r *reader) nodeServiceAuth(key string) string {
	value := strings.ToLower(r.withDefault(key, defaultNodeServiceAuth))

	if value != NodeServiceAuthNone && value != NodeServiceAuthIAM {
		r.reject(key, fmt.Sprintf("must be %q or %q", NodeServiceAuthNone, NodeServiceAuthIAM))
		return defaultNodeServiceAuth
	}

	return value
}

func (r *reader) reject(key, problem string) {
	r.problems = append(r.problems, key+" "+problem)
}
