package rate_limiter

import (
	"testing"
	"time"
)

func TestLimiterAllowAndRefill(t *testing.T) {
	l := NewLimiter(2, 10)

	if !l.Allow("127.0.0.1") {
		t.Fatal("expected first request to pass")
	}
	if !l.Allow("127.0.0.1") {
		t.Fatal("expected second request to pass")
	}
	if l.Allow("127.0.0.1") {
		t.Fatal("expected third request to be limited")
	}

	time.Sleep(150 * time.Millisecond)

	if !l.Allow("127.0.0.1") {
		t.Fatal("expected request to pass after refill")
	}
}
