package matrix_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
)

func TestNewRejectsMatricesThatBreakADomainRule(t *testing.T) {
	cases := []struct {
		name   string
		values [][]float64
	}{
		{"nil matrix", nil},
		{"matrix without rows", [][]float64{}},
		{"row without columns", [][]float64{{}}},
		{"rows of different length", [][]float64{{1, 2}, {3}}},
		{"value is NaN", [][]float64{{1, 2}, {3, math.NaN()}}},
		{"value is positive infinity", [][]float64{{math.Inf(1)}}},
		{"value is negative infinity", [][]float64{{math.Inf(-1)}}},
		{"more columns than rows", [][]float64{{1, 2, 3}}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := matrix.New(testCase.values)

			var validation *matrix.ValidationError
			require.ErrorAs(t, err, &validation)
			assert.NotEmpty(t, validation.Reason, "the reason is what reaches the client")
		})
	}
}

func TestNewAcceptsEveryValidShape(t *testing.T) {
	cases := []struct {
		name            string
		values          [][]float64
		expectedRows    int
		expectedColumns int
	}{
		{"single value", [][]float64{{5}}, 1, 1},
		{"square", [][]float64{{1, 2}, {3, 4}}, 2, 2},
		{"more rows than columns", [][]float64{{1}, {2}, {3}}, 3, 1},
		{"negative and decimal values", [][]float64{{-1.5, 0}, {0, 2.25}}, 2, 2},
		{"repeated values", [][]float64{{7, 7}, {7, 7}}, 2, 2},
		{"large values", [][]float64{{1e12, -1e12}, {1e12, 1e12}}, 2, 2},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := matrix.New(testCase.values)

			require.NoError(t, err)
			assert.Equal(t, testCase.expectedRows, result.Rows())
			assert.Equal(t, testCase.expectedColumns, result.Columns())
		})
	}
}

func TestValidationErrorMessageIsTheReason(t *testing.T) {
	_, err := matrix.New([][]float64{{1, 2}, {3}})

	var validation *matrix.ValidationError
	require.ErrorAs(t, err, &validation)
	assert.Equal(t, validation.Reason, validation.Error())
}

func TestColumnsOfAnEmptyMatrixIsZero(t *testing.T) {
	assert.Equal(t, 0, matrix.Matrix{}.Columns())
	assert.Equal(t, 0, matrix.Matrix{}.Rows())
}
