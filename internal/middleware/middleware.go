package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/sYanXO/http-server-scratch/internal/metrics"
	rate_limiter "github.com/sYanXO/http-server-scratch/internal/rate-limiter"
)

func RateLimitMiddleware(l rate_limiter.Limiter, failOpen bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "unable to get client ip", http.StatusInternalServerError)
				return
			}

			// CRITICAL: Add a strict timeout so a slow Redis doesn't hang your server
			ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
			defer cancel()

			allowed, err := l.Allow(ctx, ip)

			if err != nil {
				// INTERVIEW GREEN FLAG: Fail-open vs Fail-closed logic
				if failOpen {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Rate limiting service unavailable", http.StatusServiceUnavailable)
				return
			}

			if !allowed {
				// THIS LINE IS WHY THE IMPORT STAYS: It actively uses the metrics package
				metrics.RateLimitRejectionsTotal.Inc()

				w.Header().Set("Retry-After", "10")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
