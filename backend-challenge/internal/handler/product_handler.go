package handler

import (
	"net/http"
	"strconv"

	"github.com/oolio-group/order-management/internal/models"
	"github.com/oolio-group/order-management/internal/service"
)

// ProductHandler handles product-related HTTP requests.
type ProductHandler struct {
	svc *service.ProductService
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// ListProducts handles GET /api/product.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.ListProducts(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

// GetProduct handles GET /api/product/{productId}.
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productId")

	// Validate product ID is a valid integer string (per OpenAPI spec).
	if _, err := strconv.ParseInt(productID, 10, 64); err != nil {
		writeError(w, models.NewBadRequestError("invalid product ID"))
		return
	}

	product, err := h.svc.GetProduct(r.Context(), productID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}
