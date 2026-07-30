package matrix

import "fmt"

type ValidationError struct {
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Reason
}

func newValidationError(format string, args ...any) *ValidationError {
	return &ValidationError{Reason: fmt.Sprintf(format, args...)}
}
