package store

import "testing"

func TestUserStoreCRUD(t *testing.T) {
	s := NewUserStore()

	id := s.Create(User{Name: "Alice"})
	if id != 1 {
		t.Fatalf("expected id 1, got %d", id)
	}

	user, ok := s.Get(id)
	if !ok {
		t.Fatal("expected user to exist")
	}
	if user.Name != "Alice" {
		t.Fatalf("expected name Alice, got %q", user.Name)
	}

	if !s.Delete(id) {
		t.Fatal("expected delete to succeed")
	}

	if _, ok := s.Get(id); ok {
		t.Fatal("expected user to be deleted")
	}

	if s.Delete(id) {
		t.Fatal("expected deleting missing user to fail")
	}
}
