package dto

import (
	"time"

	"github.com/detecta/reto-tecnico/go-api/internal/application/auth"
	"github.com/detecta/reto-tecnico/go-api/internal/application/factorization"
	"github.com/detecta/reto-tecnico/go-api/internal/domain/statistics"
)

const successMessage = "Matrix processed successfully"

type FactorizationResponseDto struct {
	Success   bool                 `json:"success" example:"true"`
	Data      FactorizationDataDto `json:"data"`
	Message   string               `json:"message" example:"Matrix processed successfully"`
	Timestamp string               `json:"timestamp" example:"2026-07-30T12:00:00Z"`
}

type FactorizationDataDto struct {
	Original   [][]float64   `json:"original"`
	Q          [][]float64   `json:"q"`
	R          [][]float64   `json:"r"`
	Statistics StatisticsDto `json:"statistics"`
}

type StatisticsDto struct {
	Q MatrixStatisticsDto `json:"q"`
	R MatrixStatisticsDto `json:"r"`
}

type MatrixStatisticsDto struct {
	Maximum    float64 `json:"max" example:"0.3162"`
	Minimum    float64 `json:"min" example:"-0.9487"`
	Average    float64 `json:"average" example:"-0.4743"`
	Sum        float64 `json:"sum" example:"-1.8974"`
	IsDiagonal bool    `json:"isDiagonal" example:"false"`
}

type AccessTokenResponseDto struct {
	Success   bool           `json:"success" example:"true"`
	Data      AccessTokenDto `json:"data"`
	Message   string         `json:"message" example:"Authentication successful"`
	Timestamp string         `json:"timestamp" example:"2026-07-30T12:00:00Z"`
}

type AccessTokenDto struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType" example:"Bearer"`
	ExpiresIn   int    `json:"expiresIn" example:"3600"`
}

type HealthResponseDto struct {
	Status string `json:"status" example:"ok"`
}

func NewFactorizationResponse(result factorization.Result) FactorizationResponseDto {
	return FactorizationResponseDto{
		Success: true,
		Data: FactorizationDataDto{
			Original:   result.Original,
			Q:          result.Orthogonal,
			R:          result.UpperTriangular,
			Statistics: newStatistics(result.Statistics),
		},
		Message:   successMessage,
		Timestamp: nowAsTimestamp(),
	}
}

func NewAccessTokenResponse(accessToken auth.AccessToken) AccessTokenResponseDto {
	return AccessTokenResponseDto{
		Success: true,
		Data: AccessTokenDto{
			AccessToken: accessToken.Raw,
			TokenType:   "Bearer",
			ExpiresIn:   int(accessToken.ExpiresIn.Seconds()),
		},
		Message:   "Authentication successful",
		Timestamp: nowAsTimestamp(),
	}
}

func newStatistics(report statistics.Report) StatisticsDto {
	return StatisticsDto{
		Q: newMatrixStatistics(report.Orthogonal),
		R: newMatrixStatistics(report.UpperTriangular),
	}
}

func newMatrixStatistics(source statistics.MatrixStatistics) MatrixStatisticsDto {
	return MatrixStatisticsDto{
		Maximum:    source.Maximum,
		Minimum:    source.Minimum,
		Average:    source.Average,
		Sum:        source.Sum,
		IsDiagonal: source.IsDiagonal,
	}
}

func nowAsTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
