package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sYanXO/http-server-scratch/internal/store"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func CreateUser(w http.ResponseWriter, r *http.Request, userStore *store.UserStore) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var user store.User
	if err := dec.Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "body must contain a single JSON object", http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	id := userStore.Create(user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":   id,
		"name": user.Name,
	})
}

func GetUser(w http.ResponseWriter, r *http.Request, userStore *store.UserStore) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	user, flag := userStore.Get(id)
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

func DeleteUser(w http.ResponseWriter, r *http.Request, userStore *store.UserStore) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if !userStore.Delete(id) {
		http.Error(
			w,
			"user not found",
			http.StatusNotFound,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
