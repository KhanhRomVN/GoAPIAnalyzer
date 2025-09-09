package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	ValidationError ErrorType = "validation_error"
	NotFoundError   ErrorType = "not_found_error"
	SystemError     ErrorType = "system_error"
	AuthError       ErrorType = "auth_error"
	RateLimitError  ErrorType = "rate_limit_error"
)

// APIError represents a structured error with type and HTTP status code
type APIError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"status_code"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *APIError {
	return &APIError{
		Type:       ValidationError,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *APIError {
	return &APIError{
		Type:       NotFoundError,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// NewSystemError creates a new system error
func NewSystemError(message string) *APIError {
	return &APIError{
		Type:       SystemError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// NewAuthError creates a new authentication error
func NewAuthError(message string) *APIError {
	return &APIError{
		Type:       AuthError,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string) *APIError {
	return &APIError{
		Type:       RateLimitError,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
	}
}

// WithDetails adds details to an existing error
func (e *APIError) WithDetails(details string) *APIError {
	e.Details = details
	return e
}

// WithStatusCode sets a custom status code
func (e *APIError) WithStatusCode(code int) *APIError {
	e.StatusCode = code
	return e
}

// Type checking functions

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type == ValidationError
	}
	return false
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type == NotFoundError
	}
	return false
}

// IsSystemError checks if the error is a system error
func IsSystemError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type == SystemError
	}
	return false
}

// IsAuthError checks if the error is an authentication error
func IsAuthError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type == AuthError
	}
	return false
}

// IsRateLimitError checks if the error is a rate limit error
func IsRateLimitError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type == RateLimitError
	}
	return false
}

// GetStatusCode returns the HTTP status code for an error
func GetStatusCode(err error) int {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode
	}
	return http.StatusInternalServerError
}

// WrapError wraps a standard error as a system error
func WrapError(err error, message string) *APIError {
	if err == nil {
		return nil
	}

	return &APIError{
		Type:       SystemError,
		Message:    message,
		Details:    err.Error(),
		StatusCode: http.StatusInternalServerError,
	}
}

// Common predefined errors
var (
	ErrInvalidInput      = NewValidationError("Invalid input provided")
	ErrResourceNotFound  = NewNotFoundError("Resource not found")
	ErrInternalServer    = NewSystemError("Internal server error")
	ErrUnauthorized      = NewAuthError("Unauthorized access")
	ErrForbidden         = NewAuthError("Forbidden access").WithStatusCode(http.StatusForbidden)
	ErrRateLimitExceeded = NewRateLimitError("Rate limit exceeded")
)
