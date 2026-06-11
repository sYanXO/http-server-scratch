package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sYanXO/http-server-scratch/internal/rate-limiter"
)

func TestRateLimitMiddleware(t *testing.T) {
	l := rate_limiter.NewLimiter(1, 0)
	wrap := RateLimitMiddleware(l)

	called := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	})

	h := wrap(next)

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:1234"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr2.Code)
	}

	if called != 1 {
		t.Fatalf("expected next handler to run once, got %d", called)
	}
}
