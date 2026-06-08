package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	store := newUserStore()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		createUser(w, r, store)
	})
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		getUser(w, r, store)
	})
	mux.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		deleteUser(w, r, store)
	})

	fmt.Println("Server started on port 8080:")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println(err)
	}
}
