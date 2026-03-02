package handler

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oolio-group/order-management/internal/repository"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	pool  *pgxpool.Pool
	bloom *repository.BloomFilterSet
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(pool *pgxpool.Pool, bloom *repository.BloomFilterSet) *HealthHandler {
	return &HealthHandler{pool: pool, bloom: bloom}
}

// Liveness handles GET /healthz — simple liveness check.
func (h *HealthHandler) Liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness handles GET /readyz — checks DB connectivity and bloom filter status.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{"status": "ok"}

	if err := h.pool.Ping(r.Context()); err != nil {
		slog.ErrorContext(r.Context(), "health check failed: database unavailable", slog.Any("error", err))
		resp["status"] = "unavailable"
		resp["database"] = "unavailable"
		writeJSON(w, http.StatusServiceUnavailable, resp)
		return
	}
	resp["database"] = "ok"

	if h.bloom != nil {
		resp["bloom_filter"] = map[string]any{
			"ready": h.bloom.IsReady(),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
