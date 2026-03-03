package middleware

import (
	"net/http"
	"os"
)

// allowedOrigins returns the set of permitted CORS origins.
// Reads from CORS_ALLOWED_ORIGINS env var (comma-separated), defaulting to localhost dev origins.
func allowedOrigins() map[string]bool {
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		origins := make(map[string]bool)
		for _, o := range splitAndTrim(v) {
			origins[o] = true
		}
		return origins
	}
	return map[string]bool{
		"http://localhost:3000": true,
		"http://localhost:3001": true,
	}
}

func splitAndTrim(s string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			// Trim spaces manually
			j, k := 0, len(part)-1
			for j < len(part) && part[j] == ' ' {
				j++
			}
			for k >= j && part[k] == ' ' {
				k--
			}
			if j <= k {
				result = append(result, part[j:k+1])
			}
			start = i + 1
		}
	}
	return result
}

// CORS returns middleware that sets CORS headers.
// Only origins in the allowed list are permitted.
func CORS() func(http.Handler) http.Handler {
	allowed := allowedOrigins()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, api_key")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
