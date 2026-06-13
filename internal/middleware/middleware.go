package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	rate_limiter "github.com/sYanXO/http-server-scratch/internal/rate-limiter"
)

// we pass the interface rate_limiter.Limiter, not a pointer *rate_limiter.Limiter
func RateLimitMiddleware(l rate_limiter.Limiter, failOpen bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "unable to get client ip", http.StatusInternalServerError)
				return
			}

			// Add a strict timeout so a slow Redis doesn't hang your server
			ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
			defer cancel()

			allowed, err := l.Allow(ctx, ip)

			if err != nil {
				// Fail-open vs Fail-closed logic
				if failOpen {
					// In production, you would log this error here using your logger.
					// We let the request through to maintain availability.
					next.ServeHTTP(w, r)
					return
				}
				// Fail-closed: Block the request if Redis is down
				http.Error(w, "Rate limiting service unavailable", http.StatusServiceUnavailable)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", "10")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
