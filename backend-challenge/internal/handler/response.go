// Package handler provides HTTP handlers for the API endpoints.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/oolio-group/order-management/internal/models"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// writeError writes an error response. If the error is an APIError, its status code is used.
func writeError(w http.ResponseWriter, err error) {
	WriteErrorPublic(w, err)
}

// WriteErrorPublic writes an error response (exported for middleware use).
func WriteErrorPublic(w http.ResponseWriter, err error) {
	if apiErr, ok := err.(*models.APIError); ok {
		writeJSON(w, apiErr.StatusCode, apiErr)
		return
	}
	writeJSON(w, http.StatusInternalServerError, models.NewServiceUnavailableError())
}
