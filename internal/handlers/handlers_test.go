package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sYanXO/http-server-scratch/internal/store"
)

func TestCreateGetDeleteUserFlow(t *testing.T) {
	userStore := store.NewUserStore()
	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		CreateUser(w, r, userStore)
	})
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		GetUser(w, r, userStore)
	})
	mux.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		DeleteUser(w, r, userStore)
	})

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice"}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var created map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	if created["name"] != "Alice" {
		t.Fatalf("expected name Alice, got %v", created["name"])
	}

	req = httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-type"), "application/json") && !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
		t.Fatal("expected JSON content type")
	}

	var user store.User
	if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
		t.Fatalf("unmarshal get response: %v", err)
	}
	if user.Name != "Alice" {
		t.Fatalf("expected name Alice, got %q", user.Name)
	}

	req = httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rr.Code)
	}
}

func TestCreateUserValidation(t *testing.T) {
	userStore := store.NewUserStore()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "bad json", body: `{"name":`, wantStatus: http.StatusBadRequest},
		{name: "missing name", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "unknown field", body: `{"name":"Alice","age":10}`, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			CreateUser(rr, req, userStore)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
