package statistics

type MatrixStatistics struct {
	Maximum    float64
	Minimum    float64
	Average    float64
	Sum        float64
	IsDiagonal bool
}

type Report struct {
	Orthogonal      MatrixStatistics
	UpperTriangular MatrixStatistics
}
