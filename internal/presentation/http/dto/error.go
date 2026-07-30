package dto

type ErrorResponseDto struct {
	Success   bool     `json:"success" example:"false"`
	Message   string   `json:"message" example:"Invalid matrix"`
	Errors    []string `json:"errors"`
	Timestamp string   `json:"timestamp" example:"2026-07-30T12:00:00Z"`
}

func NewErrorResponse(message string, details ...string) ErrorResponseDto {
	if details == nil {
		details = []string{}
	}
	return ErrorResponseDto{
		Success:   false,
		Message:   message,
		Errors:    details,
		Timestamp: nowAsTimestamp(),
	}
}
