package middleware

import (
	"net/http"
	"runtime/debug"

	"geotrack/backend/pkg/logger"
)

// Recoverer intercepts panics in HTTP handlers, logs the stack trace, and writes a standard 500 JSON error
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				reqID := r.Header.Get("X-Request-ID")
				logger.Error().
					Str("request_id", reqID).
					Interface("panic", rvr).
					Bytes("stack", debug.Stack()).
					Msg("Unhandled panic recovered in HTTP pipeline")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"success":false,"error":{"code":"INTERNAL_SERVER_ERROR","message":"An unexpected internal server error occurred."}}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

