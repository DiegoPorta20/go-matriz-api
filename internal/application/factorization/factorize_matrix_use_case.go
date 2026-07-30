package factorization

import (
	"context"
	"fmt"

	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/statistics"
)

type FactorizeMatrixUseCase struct {
	factorizer         QRFactorizer
	statisticsProvider StatisticsProvider
	maxDimension       int
}

func NewFactorizeMatrixUseCase(
	factorizer QRFactorizer,
	statisticsProvider StatisticsProvider,
	maxDimension int,
) *FactorizeMatrixUseCase {
	return &FactorizeMatrixUseCase{
		factorizer:         factorizer,
		statisticsProvider: statisticsProvider,
		maxDimension:       maxDimension,
	}
}

type Result struct {
	Original        matrix.Matrix
	Orthogonal      matrix.Matrix
	UpperTriangular matrix.Matrix
	Statistics      statistics.Report
}

func (uc *FactorizeMatrixUseCase) Execute(ctx context.Context, values [][]float64) (Result, error) {
	original, err := matrix.New(values)
	if err != nil {
		return Result{}, err
	}

	if err := uc.enforceDimensionLimit(original); err != nil {
		return Result{}, err
	}

	orthogonal, upperTriangular, err := uc.factorizer.Factorize(original)
	if err != nil {
		return Result{}, fmt.Errorf("factorize matrix: %w", err)
	}

	report, err := uc.statisticsProvider.Calculate(ctx, orthogonal, upperTriangular)
	if err != nil {
		return Result{}, fmt.Errorf("calculate statistics: %w", err)
	}

	return Result{
		Original:        original,
		Orthogonal:      orthogonal,
		UpperTriangular: upperTriangular,
		Statistics:      report,
	}, nil
}

func (uc *FactorizeMatrixUseCase) enforceDimensionLimit(input matrix.Matrix) error {
	if input.Rows() <= uc.maxDimension && input.Columns() <= uc.maxDimension {
		return nil
	}
	return &matrix.ValidationError{Reason: fmt.Sprintf(
		"Matrix must not exceed %d rows or columns, but it has %d rows and %d columns",
		uc.maxDimension, input.Rows(), input.Columns())}
}
