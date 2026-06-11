package middleware

import (
	"net"
	"net/http"

	"github.com/sYanXO/http-server-scratch/internal/rate-limiter"
)

func RateLimitMiddleware(l *rate_limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "unable to get client ip", http.StatusInternalServerError)
				return
			}

			if !l.Allow(ip) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
