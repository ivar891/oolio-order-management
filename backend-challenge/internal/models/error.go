package models

import (
	"fmt"
	"net/http"
)

// APIError represents a structured error response.
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewBadRequestError creates a 400 Bad Request error.
func NewBadRequestError(message string) *APIError {
	return &APIError{Code: "bad_request", Message: message, StatusCode: http.StatusBadRequest}
}

// NewUnauthorizedError creates a 401 Unauthorized error.
func NewUnauthorizedError(message string) *APIError {
	return &APIError{Code: "unauthorized", Message: message, StatusCode: http.StatusUnauthorized}
}

// NewForbiddenError creates a 403 Forbidden error.
func NewForbiddenError(message string) *APIError {
	return &APIError{Code: "forbidden", Message: message, StatusCode: http.StatusForbidden}
}

// NewNotFoundError creates a 404 Not Found error.
func NewNotFoundError(message string) *APIError {
	return &APIError{Code: "not_found", Message: message, StatusCode: http.StatusNotFound}
}

// NewValidationError creates a 422 Validation error.
func NewValidationError(message string) *APIError {
	return &APIError{Code: "validation", Message: message, StatusCode: http.StatusUnprocessableEntity}
}

// NewConstraintError creates a 422 Constraint error.
func NewConstraintError(message string) *APIError {
	return &APIError{Code: "constraint", Message: message, StatusCode: http.StatusUnprocessableEntity}
}

// NewInsufficientStockError creates a 422 error for insufficient stock.
func NewInsufficientStockError(productName string, available, requested int) *APIError {
	msg := fmt.Sprintf("insufficient stock for product: %s (available: %d, requested: %d)", productName, available, requested)
	return &APIError{Code: "insufficient_stock", Message: msg, StatusCode: http.StatusUnprocessableEntity}
}

// NewUnsupportedMediaTypeError creates a 415 error.
func NewUnsupportedMediaTypeError() *APIError {
	return &APIError{Code: "unsupported_media_type", Message: "Content-Type must be application/json", StatusCode: http.StatusUnsupportedMediaType}
}

// NewServiceUnavailableError creates a 503 error.
func NewServiceUnavailableError() *APIError {
	return &APIError{Code: "unavailable", Message: "service temporarily unavailable", StatusCode: http.StatusServiceUnavailable}
}
