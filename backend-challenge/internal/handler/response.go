// Package handler provides HTTP handlers for the API endpoints.
package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/oolio-group/order-management/internal/models"
)

// writeJSON writes a JSON response with the given status code.
// It buffers the JSON encoding to avoid writing headers before knowing
// whether encoding succeeds.
func writeJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		slog.Error("Failed to encode JSON response", slog.Any("error", err))
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// writeError writes an error response. If the error is an APIError, its status code is used.
func writeError(w http.ResponseWriter, err error) {
	WriteErrorPublic(w, err)
}

// WriteErrorPublic writes an error response (exported for middleware use).
func WriteErrorPublic(w http.ResponseWriter, err error) {
	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		writeJSON(w, apiErr.StatusCode, apiErr)
		return
	}
	writeJSON(w, http.StatusInternalServerError, models.NewServiceUnavailableError())
}
