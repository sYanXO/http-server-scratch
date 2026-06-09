package main

import (
	"fmt"
	"net/http"
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

	fmt.Println("Server started on port 8080:")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println(err)
	}
}
