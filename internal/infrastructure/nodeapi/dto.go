package nodeapi

import (
	"github.com/detecta/reto-tecnico/go-api/internal/domain/matrix"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/statistics"
)

type statisticsRequest struct {
	Q matrix.Matrix `json:"q"`
	R matrix.Matrix `json:"r"`
}

type statisticsResponse struct {
	Success bool                   `json:"success"`
	Data    statisticsResponseData `json:"data"`
}

type statisticsResponseData struct {
	Q matrixStatistics `json:"q"`
	R matrixStatistics `json:"r"`
}

type matrixStatistics struct {
	Maximum    float64 `json:"max"`
	Minimum    float64 `json:"min"`
	Average    float64 `json:"average"`
	Sum        float64 `json:"sum"`
	IsDiagonal bool    `json:"isDiagonal"`
}

func (r statisticsResponse) toReport() statistics.Report {
	return statistics.Report{
		Orthogonal:      r.Data.Q.toStatistics(),
		UpperTriangular: r.Data.R.toStatistics(),
	}
}

func (s matrixStatistics) toStatistics() statistics.MatrixStatistics {
	return statistics.MatrixStatistics{
		Maximum:    s.Maximum,
		Minimum:    s.Minimum,
		Average:    s.Average,
		Sum:        s.Sum,
		IsDiagonal: s.IsDiagonal,
	}
}
