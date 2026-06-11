package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	store := newUserStore()

	limiter := NewLimiter(10, 1)
	wrap := rateLimitMiddleware(limiter)

	mux.Handle("/", wrap(http.HandlerFunc(handleRoot)))

	mux.Handle("POST /users", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createUser(w, r, store)
	})))

	mux.Handle("GET /users/{id}", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getUser(w, r, store)
	})))

	mux.Handle("DELETE /users/{id}", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deleteUser(w, r, store)
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
