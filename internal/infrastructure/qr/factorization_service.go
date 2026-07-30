package qr

import (
	"fmt"

	"gonum.org/v1/gonum/mat"

	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
)

type FactorizationService struct{}

func NewFactorizationService() FactorizationService {
	return FactorizationService{}
}

func (FactorizationService) Factorize(input matrix.Matrix) (matrix.Matrix, matrix.Matrix, error) {
	if input.Rows() < input.Columns() {
		return nil, nil, fmt.Errorf(
			"qr factorization needs at least as many rows as columns, got %d×%d",
			input.Rows(), input.Columns())
	}

	var factorization mat.QR
	factorization.Factorize(toDense(input))

	var orthogonal, upperTriangular mat.Dense
	factorization.QTo(&orthogonal)
	factorization.RTo(&upperTriangular)

	return fromDense(&orthogonal), fromDense(&upperTriangular), nil
}

func toDense(input matrix.Matrix) *mat.Dense {
	rows, columns := input.Rows(), input.Columns()

	flattened := make([]float64, 0, rows*columns)
	for _, row := range input {
		flattened = append(flattened, row...)
	}

	return mat.NewDense(rows, columns, flattened)
}

func fromDense(source *mat.Dense) matrix.Matrix {
	rows, columns := source.Dims()

	result := make(matrix.Matrix, rows)
	for row := range rows {
		result[row] = make([]float64, columns)
		for column := range columns {
			result[row][column] = source.At(row, column)
		}
	}

	return result
}
