package rulesengine

import "fmt"

// UndefinedFactError represents an error for an undefined fact
type UndefinedFactError struct {
	Message string
	Code    string
}

// Error implements the error interface for UndefinedFactError
func (e *UndefinedFactError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewUndefinedFactError creates a new UndefinedFactError instance
func NewUndefinedFactError(message string) *UndefinedFactError {
	return &UndefinedFactError{
		Message: message,
		Code:    "UNDEFINED_FACT",
	}
}
