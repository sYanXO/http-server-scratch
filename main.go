package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	Name string `json:"name"`
}

var userDB = make(map[int]User)
var userMutex sync.RWMutex

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)

	mux.HandleFunc("POST /users", createUser)

	mux.HandleFunc("GET /users/{id}", getUser)

	mux.HandleFunc("DELETE /users/{id}", deleteUser)

	fmt.Println("Server started on port 8080:")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if user.Name == "" {
		http.Error(w,
			"name is required",
			http.StatusBadRequest,
		)

		return
	}

	userMutex.Lock()
	defer userMutex.Unlock()

	userDB[len(userDB)+1] = user
	w.WriteHeader(http.StatusNoContent)

}

func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	userMutex.RLock()
	user, flag := userDB[id]
	userMutex.RUnlock()

	if !flag {
		http.Error(w,
			"user not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-type", "application/json")

	j, err := json.Marshal(user)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if _, ok := userDB[id]; !ok {
		http.Error(
			w,
			"user not found",
			http.StatusBadRequest,
		)
		return
	}
	userMutex.Lock()
	delete(userDB, id)
	userMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
