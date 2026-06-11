package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sYanXO/http-server-scratch/internal/handlers"
	"github.com/sYanXO/http-server-scratch/internal/middleware"
	"github.com/sYanXO/http-server-scratch/internal/rate-limiter"
	"github.com/sYanXO/http-server-scratch/internal/store"
)

func TestServerWiring(t *testing.T) {
	userStore := store.NewUserStore()
	limiter := rate_limiter.NewLimiter(10, 1)
	wrap := middleware.RateLimitMiddleware(limiter)

	mux := http.NewServeMux()
	mux.Handle("/", wrap(http.HandlerFunc(handlers.HandleRoot)))
	mux.Handle("POST /users", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUser(w, r, userStore)
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "hello world") {
		t.Fatalf("expected root response, got %q", rr.Body.String())
	}
}
