package factorization

import (
	"context"
	"errors"

	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/statistics"
)

var ErrStatisticsUnavailable = errors.New("statistics service unavailable")

type QRFactorizer interface {
	Factorize(input matrix.Matrix) (orthogonal, upperTriangular matrix.Matrix, err error)
}

type StatisticsProvider interface {
	Calculate(ctx context.Context, orthogonal, upperTriangular matrix.Matrix) (statistics.Report, error)
}
