package factorization

import (
	"context"
	"errors"

	"github.com/DiegoPorta20/go-matriz-api/internal/domain/matrix"
	"github.com/DiegoPorta20/go-matriz-api/internal/domain/statistics"
)

var ErrStatisticsUnavailable = errors.New("statistics service unavailable")

type QRFactorizer interface {
	Factorize(input matrix.Matrix) (orthogonal, upperTriangular matrix.Matrix, err error)
}

type StatisticsProvider interface {
	Calculate(ctx context.Context, orthogonal, upperTriangular matrix.Matrix) (statistics.Report, error)
}
