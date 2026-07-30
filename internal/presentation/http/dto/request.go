package dto

type MatrixRequestDto struct {
	Matrix [][]float64 `json:"matrix"`
}

type LoginRequestDto struct {
	Username string `json:"username" example:"demo"`
	Password string `json:"password" example:"change-this-password"`
}
