package models

import (
	"net/http"
	"testing"
)

func TestAPIErrorImplementsError(t *testing.T) {
	err := NewBadRequestError("test")
	if err.Error() != "bad_request: test" {
		t.Errorf("unexpected error string: %s", err.Error())
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name       string
		err        *APIError
		code       string
		statusCode int
	}{
		{"bad_request", NewBadRequestError("msg"), "bad_request", http.StatusBadRequest},
		{"unauthorized", NewUnauthorizedError("msg"), "unauthorized", http.StatusUnauthorized},
		{"forbidden", NewForbiddenError("msg"), "forbidden", http.StatusForbidden},
		{"not_found", NewNotFoundError("msg"), "not_found", http.StatusNotFound},
		{"validation", NewValidationError("msg"), "validation", http.StatusUnprocessableEntity},
		{"constraint", NewConstraintError("msg"), "constraint", http.StatusUnprocessableEntity},
		{"unsupported_media_type", NewUnsupportedMediaTypeError(), "unsupported_media_type", http.StatusUnsupportedMediaType},
		{"unavailable", NewServiceUnavailableError(), "unavailable", http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("code: got %s, want %s", tt.err.Code, tt.code)
			}
			if tt.err.StatusCode != tt.statusCode {
				t.Errorf("status: got %d, want %d", tt.err.StatusCode, tt.statusCode)
			}
		})
	}
}

func TestInsufficientStockError(t *testing.T) {
	err := NewInsufficientStockError("Waffle", 1, 5)
	if err.Code != "insufficient_stock" {
		t.Errorf("expected code insufficient_stock, got %s", err.Code)
	}
	if err.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", err.StatusCode)
	}
}
