package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"geotrack/backend/pkg/logger"
)

// StructuredLogger returns a middleware that logs incoming HTTP requests using Zerolog
func StructuredLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				reqID := middleware.GetReqID(r.Context())
				duration := time.Since(t1)

				logger.Info().
					Str("request_id", reqID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("remote_ip", r.RemoteAddr).
					Int("status", ww.Status()).
					Int("bytes_written", ww.BytesWritten()).
					Dur("duration_ms", duration).
					Msg("HTTP Request")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// SimpleCORS middleware handles basic CORS headers
func SimpleCORS(allowedOrigins []string) func(next http.Handler) http.Handler {
	originsMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		originsMap[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if originsMap[origin] || len(allowedOrigins) == 0 {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "300")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

