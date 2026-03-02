package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/oolio-group/order-management/internal/handler"
	"github.com/oolio-group/order-management/internal/models"
)

// Recovery returns middleware that recovers from panics.
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					slog.ErrorContext(r.Context(), "panic recovered",
						slog.Any("panic", rec),
						slog.String("stack", string(debug.Stack())),
					)
					handler.WriteErrorPublic(w, models.NewServiceUnavailableError())
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
