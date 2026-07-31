package factorization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DiegoPorta20/go-matriz-api/internal/application/factorization"
	"github.com/DiegoPorta20/go-matriz-api/internal/domain/matrix"
	"github.com/DiegoPorta20/go-matriz-api/internal/domain/statistics"
)

const maxDimension = 3

type stubFactorizer struct {
	orthogonal      matrix.Matrix
	upperTriangular matrix.Matrix
	err             error
	receivedInput   matrix.Matrix
	calls           int
}

func (s *stubFactorizer) Factorize(input matrix.Matrix) (matrix.Matrix, matrix.Matrix, error) {
	s.calls++
	s.receivedInput = input
	return s.orthogonal, s.upperTriangular, s.err
}

type stubStatisticsProvider struct {
	report    statistics.Report
	err       error
	receivedQ matrix.Matrix
	receivedR matrix.Matrix
	calls     int
}

func (s *stubStatisticsProvider) Calculate(
	_ context.Context,
	orthogonal, upperTriangular matrix.Matrix,
) (statistics.Report, error) {
	s.calls++
	s.receivedQ, s.receivedR = orthogonal, upperTriangular
	return s.report, s.err
}

func TestExecuteReturnsTheConsolidatedResult(t *testing.T) {
	factorizer := &stubFactorizer{
		orthogonal:      matrix.Matrix{{1, 0}, {0, 1}},
		upperTriangular: matrix.Matrix{{1, 2}, {0, 4}},
	}
	provider := &stubStatisticsProvider{report: statistics.Report{
		Orthogonal:      statistics.MatrixStatistics{Maximum: 1, Sum: 2, IsDiagonal: true},
		UpperTriangular: statistics.MatrixStatistics{Maximum: 4, Sum: 7},
	}}
	useCase := factorization.NewFactorizeMatrixUseCase(factorizer, provider, maxDimension)

	result, err := useCase.Execute(context.Background(), [][]float64{{1, 2}, {3, 4}})

	require.NoError(t, err)
	assert.Equal(t, matrix.Matrix{{1, 2}, {3, 4}}, result.Original)
	assert.Equal(t, factorizer.orthogonal, result.Orthogonal)
	assert.Equal(t, factorizer.upperTriangular, result.UpperTriangular)
	assert.True(t, result.Statistics.Orthogonal.IsDiagonal)
	assert.InDelta(t, 7.0, result.Statistics.UpperTriangular.Sum, 0)
}

func TestExecutePassesTheFactorsToTheStatisticsProvider(t *testing.T) {
	factorizer := &stubFactorizer{
		orthogonal:      matrix.Matrix{{1}},
		upperTriangular: matrix.Matrix{{9}},
	}
	provider := &stubStatisticsProvider{}
	useCase := factorization.NewFactorizeMatrixUseCase(factorizer, provider, maxDimension)

	_, err := useCase.Execute(context.Background(), [][]float64{{9}})

	require.NoError(t, err)
	assert.Equal(t, matrix.Matrix{{9}}, factorizer.receivedInput)
	assert.Equal(t, factorizer.orthogonal, provider.receivedQ)
	assert.Equal(t, factorizer.upperTriangular, provider.receivedR)
}

func TestExecuteRejectsAnInvalidMatrixWithoutCallingAnything(t *testing.T) {
	factorizer := &stubFactorizer{}
	provider := &stubStatisticsProvider{}
	useCase := factorization.NewFactorizeMatrixUseCase(factorizer, provider, maxDimension)

	_, err := useCase.Execute(context.Background(), [][]float64{{1, 2}, {3}})

	var validation *matrix.ValidationError
	require.ErrorAs(t, err, &validation)
	assert.Zero(t, factorizer.calls, "an invalid matrix must not reach the factorizer")
	assert.Zero(t, provider.calls)
}

func TestExecuteRejectsAMatrixLargerThanTheConfiguredLimit(t *testing.T) {
	factorizer := &stubFactorizer{}
	useCase := factorization.NewFactorizeMatrixUseCase(factorizer, &stubStatisticsProvider{}, maxDimension)
	oversized := [][]float64{{1, 2, 3, 4}, {1, 2, 3, 4}, {1, 2, 3, 4}, {1, 2, 3, 4}}

	_, err := useCase.Execute(context.Background(), oversized)

	var validation *matrix.ValidationError
	require.ErrorAs(t, err, &validation)
	assert.Contains(t, validation.Reason, "3")
	assert.Zero(t, factorizer.calls)
}

func TestExecutePropagatesAFactorizationFailure(t *testing.T) {
	failure := errors.New("lapack exploded")
	provider := &stubStatisticsProvider{}
	useCase := factorization.NewFactorizeMatrixUseCase(
		&stubFactorizer{err: failure}, provider, maxDimension)

	_, err := useCase.Execute(context.Background(), [][]float64{{1}})

	require.ErrorIs(t, err, failure)
	assert.Zero(t, provider.calls, "statistics must not be requested for a failed factorization")
}

func TestExecutePreservesTheUnavailableStatisticsError(t *testing.T) {
	useCase := factorization.NewFactorizeMatrixUseCase(
		&stubFactorizer{orthogonal: matrix.Matrix{{1}}, upperTriangular: matrix.Matrix{{1}}},
		&stubStatisticsProvider{err: factorization.ErrStatisticsUnavailable},
		maxDimension,
	)

	_, err := useCase.Execute(context.Background(), [][]float64{{1}})

	require.ErrorIs(t, err, factorization.ErrStatisticsUnavailable)
}
