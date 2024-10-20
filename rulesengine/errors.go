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

// InvalidRuleError represents an error for an invalid rule
type InvalidRuleError struct {
	Message string
	Code    string
}

func (e *InvalidRuleError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewInvalidRuleError(message string, code string) *InvalidRuleError {
	return &InvalidRuleError{
		Message: message,
		Code:    code,
	}
}

func NewInvalidPriorityTypeError() *InvalidRuleError {
	return NewInvalidRuleError("Priority must be an integer", "INVALID_PRIORITY_TYPE")
}

func NewInvalidPriorityValueError() *InvalidRuleError {
	return NewInvalidRuleError("Priority must be greater than zero", "INVALID_PRIORITY_VALUE")
}

func NewPriorityNotSetError() *InvalidRuleError {
	return NewInvalidRuleError("Priority not set", "PRIORITY_NOT_SET")
}
