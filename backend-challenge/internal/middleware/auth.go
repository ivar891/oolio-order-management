// Package middleware provides HTTP middleware for the API.
package middleware

import (
	"net/http"
	"strings"

	"github.com/oolio-group/order-management/internal/handler"
	"github.com/oolio-group/order-management/internal/models"
)

// Auth returns middleware that validates the api_key header.
func Auth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := strings.TrimSpace(r.Header.Get("api_key"))
			if key == "" {
				handler.WriteErrorPublic(w, models.NewUnauthorizedError("API key is required"))
				return
			}
			if key != apiKey {
				handler.WriteErrorPublic(w, models.NewForbiddenError("invalid API key"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
