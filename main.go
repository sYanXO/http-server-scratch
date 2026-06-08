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

type userStore struct {
	mu     sync.RWMutex
	users  map[int]User
	nextID int
}

func newUserStore() *userStore {
	return &userStore{
		users:  make(map[int]User),
		nextID: 1,
	}
}

func (s *userStore) Create(user User) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.users[id] = user
	s.nextID++
	return id
}

func (s *userStore) Get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	return user, ok
}

func (s *userStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return false
	}
	delete(s.users, id)
	return true
}

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

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func createUser(w http.ResponseWriter, r *http.Request, store *userStore) {
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

	store.Create(user)
	w.WriteHeader(http.StatusNoContent)

}

func getUser(w http.ResponseWriter, r *http.Request, store *userStore) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	user, flag := store.Get(id)
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

func deleteUser(w http.ResponseWriter, r *http.Request, store *userStore) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if !store.Delete(id) {
		http.Error(
			w,
			"user not found",
			http.StatusNotFound,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
