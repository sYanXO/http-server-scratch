package main

import (
	"sync"
	"time"
)

type Bucket struct {
	tokens         int
	lastRefillTime time.Time
}

type Limiter struct {
	mu         sync.Mutex
	buckets    map[string]*Bucket
	capacity   int
	refillRate float64
}

func NewLimiter(capacity int, refillRate float64) *Limiter {
	return &Limiter{
		buckets:    make(map[string]*Bucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}
func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[ip]
	if !ok {
		l.buckets[ip] = &Bucket{
			tokens:         l.capacity - 1,
			lastRefillTime: now,
		}
		return true
	}
	elapsed := now.Sub(b.lastRefillTime).Seconds()
	refilled := int(elapsed * l.refillRate)

	if refilled > 0 {
		b.tokens += refilled
		if b.tokens > l.capacity {
			b.tokens = l.capacity
		}
		b.lastRefillTime = now
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true

}
