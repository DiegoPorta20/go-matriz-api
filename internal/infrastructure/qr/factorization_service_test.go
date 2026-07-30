package qr_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
	"github.com/detecta/reto-tecnico/go-api/internal/infrastructure/qr"
)

const tolerance = 1e-9

func TestFactorizeProducesAValidFactorization(t *testing.T) {
	cases := []struct {
		name  string
		input matrix.Matrix
	}{
		{"single value", matrix.Matrix{{5}}},
		{"square", matrix.Matrix{{1, 2}, {3, 4}}},
		{"more rows than columns", matrix.Matrix{{1, 2}, {3, 4}, {5, 6}}},
		{"single column", matrix.Matrix{{3}, {4}}},
		{"already diagonal", matrix.Matrix{{2, 0}, {0, 3}}},
		{"negative and decimal values", matrix.Matrix{{-1.5, 2.25}, {0.5, -3}}},
		{"singular", matrix.Matrix{{1, 2}, {2, 4}}},
		{"large values", matrix.Matrix{{1e8, 1}, {1, 1e8}}},
	}

	service := qr.NewFactorizationService()

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			rows, columns := testCase.input.Rows(), testCase.input.Columns()

			orthogonal, upperTriangular, err := service.Factorize(testCase.input)
			require.NoError(t, err)

			assert.Equal(t, rows, orthogonal.Rows(), "Q must be m×m")
			assert.Equal(t, rows, orthogonal.Columns(), "Q must be m×m")
			assert.Equal(t, rows, upperTriangular.Rows(), "R must be m×n")
			assert.Equal(t, columns, upperTriangular.Columns(), "R must be m×n")

			assertMatricesAreClose(t, testCase.input, multiply(orthogonal, upperTriangular), "Q*R must equal the input")
			assertMatricesAreClose(t, identity(rows), multiply(transpose(orthogonal), orthogonal), "Q must be orthogonal")
			assertUpperTriangular(t, upperTriangular)
		})
	}
}

func TestFactorizeRejectsMoreColumnsThanRows(t *testing.T) {
	_, _, err := qr.NewFactorizationService().Factorize(matrix.Matrix{{1, 2, 3}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rows")
}

func multiply(left, right matrix.Matrix) matrix.Matrix {
	result := make(matrix.Matrix, left.Rows())
	for row := range left.Rows() {
		result[row] = make([]float64, right.Columns())
		for column := range right.Columns() {
			sum := 0.0
			for index := range left.Columns() {
				sum += left[row][index] * right[index][column]
			}
			result[row][column] = sum
		}
	}
	return result
}

func transpose(source matrix.Matrix) matrix.Matrix {
	result := make(matrix.Matrix, source.Columns())
	for row := range source.Columns() {
		result[row] = make([]float64, source.Rows())
		for column := range source.Rows() {
			result[row][column] = source[column][row]
		}
	}
	return result
}

func identity(size int) matrix.Matrix {
	result := make(matrix.Matrix, size)
	for row := range size {
		result[row] = make([]float64, size)
		result[row][row] = 1
	}
	return result
}

func assertMatricesAreClose(t *testing.T, expected, actual matrix.Matrix, message string) {
	t.Helper()

	require.Equal(t, expected.Rows(), actual.Rows(), message)
	require.Equal(t, expected.Columns(), actual.Columns(), message)

	for row := range expected.Rows() {
		for column := range expected.Columns() {
			assert.InDelta(t, expected[row][column], actual[row][column],
				relativeTolerance(expected[row][column]), "%s at [%d][%d]", message, row, column)
		}
	}
}

func relativeTolerance(expected float64) float64 {
	return tolerance * math.Max(1, math.Abs(expected))
}

func assertUpperTriangular(t *testing.T, candidate matrix.Matrix) {
	t.Helper()

	for row := range candidate.Rows() {
		for column := range min(row, candidate.Columns()) {
			assert.InDelta(t, 0.0, candidate[row][column], tolerance,
				"R must be upper triangular, but [%d][%d] is not zero", row, column)
		}
	}
}
