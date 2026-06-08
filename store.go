package main

import "sync"

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
