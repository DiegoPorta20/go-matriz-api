package matrix

import "math"

type Matrix [][]float64

func New(values [][]float64) (Matrix, error) {
	if len(values) == 0 {
		return nil, newValidationError("Matrix must contain at least one row")
	}

	columns := len(values[0])
	if columns == 0 {
		return nil, newValidationError("Matrix must contain at least one column")
	}

	for index, row := range values {
		if len(row) != columns {
			return nil, newValidationError(
				"All matrix rows must have %d columns, but row %d has %d",
				columns, index+1, len(row))
		}
		if err := validateFiniteValues(row, index); err != nil {
			return nil, err
		}
	}

	if len(values) < columns {
		return nil, newValidationError(
			"Matrix must have at least as many rows as columns, but it has %d rows and %d columns",
			len(values), columns)
	}

	return Matrix(values), nil
}

func (m Matrix) Rows() int {
	return len(m)
}

func (m Matrix) Columns() int {
	if len(m) == 0 {
		return 0
	}
	return len(m[0])
}

func validateFiniteValues(row []float64, rowIndex int) error {
	for columnIndex, value := range row {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return newValidationError(
				"Matrix value at row %d, column %d must be a finite number",
				rowIndex+1, columnIndex+1)
		}
	}
	return nil
}
