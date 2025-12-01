package domain

import "errors"

// Domain errors
var (
	// URL errors
	ErrURLNotFound     = errors.New("URL not found")
	ErrURLExpired      = errors.New("URL has expired")
	ErrURLInactive     = errors.New("URL is inactive")
	ErrDuplicateShortCode = errors.New("short code already exists")
	
	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
	
	// Security errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrForbidden         = errors.New("forbidden")
	
	// Storage errors
	ErrStorageFailure    = errors.New("storage operation failed")
	ErrCacheFailure      = errors.New("cache operation failed")
	
	// Service errors
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
)

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(code int, message string, err error) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

