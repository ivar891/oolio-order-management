package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/oolio-group/order-management/internal/models"
	"github.com/oolio-group/order-management/internal/service"
)

const maxRequestBodySize = 1 << 20 // 1MB

// OrderHandler handles order-related HTTP requests.
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// PlaceOrder handles POST /api/order.
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// Validate Content-Type.
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		writeError(w, models.NewUnsupportedMediaTypeError())
		return
	}

	// Limit request body size.
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, models.NewBadRequestError("request body too large"))
			return
		}
		writeError(w, models.NewBadRequestError("failed to read request body"))
		return
	}

	var req models.OrderRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, models.NewBadRequestError("invalid JSON body"))
		return
	}

	resp, err := h.svc.PlaceOrder(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
