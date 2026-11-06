package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*clientVisitor)
	mu       sync.Mutex
)

func init() {
	// Periodic cleanup of stale visitor records every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// RateLimit limits requests per IP using a token bucket algorithm
func RateLimit(requestsPerMin int, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMin)), burst)
				v = &clientVisitor{limiter: limiter, lastSeen: time.Now()}
				visitors[ip] = v
			}
			v.lastSeen = time.Now()
			allowed := v.limiter.Allow()
			mu.Unlock()

			if !allowed {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"success":false,"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests. Please retry in 60 seconds."}}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

