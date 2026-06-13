package store

import "context"

type User struct {
	Name string `json:"name"`
}

type Limiter interface {
	Allow(ctx context.Context, identifyer string) (bool, error)
}
