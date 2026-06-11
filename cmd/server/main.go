package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sYanXO/http-server-scratch/internal/handlers"
	"github.com/sYanXO/http-server-scratch/internal/middleware"
	rate_limiter "github.com/sYanXO/http-server-scratch/internal/rate-limiter"
	"github.com/sYanXO/http-server-scratch/internal/store"
)

func main() {
	mux := http.NewServeMux()
	userStore := store.NewUserStore()

	limiter := rate_limiter.NewLimiter(10, 1)
	wrap := middleware.RateLimitMiddleware(limiter)

	mux.Handle("/", wrap(http.HandlerFunc(handlers.HandleRoot)))

	mux.Handle("POST /users", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUser(w, r, userStore)
	})))

	mux.Handle("GET /users/{id}", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.GetUser(w, r, userStore)
	})))

	mux.Handle("DELETE /users/{id}", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteUser(w, r, userStore)
	})))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		fmt.Println("Server started on port 8080.")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println(("Shutting down server...."))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Shutdown error: ", err)
	} else {
		fmt.Println("Server gracefully stopped")
	}
}
