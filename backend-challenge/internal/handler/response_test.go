package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oolio-group/order-management/internal/models"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestWriteErrorAPIError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, models.NewNotFoundError("not found"))

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
	var resp models.APIError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code != "not_found" {
		t.Errorf("expected code 'not_found', got %s", resp.Code)
	}
}

func TestWriteErrorGeneric(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, fmt.Errorf("generic error"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}
