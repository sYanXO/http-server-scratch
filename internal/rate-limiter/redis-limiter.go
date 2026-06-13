package rate_limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter defines the contract for rate limiting.
// Using an interface allows us to easily swap between Redis and in-memory for testing.
type Limiter interface {
	Allow(ctx context.Context, identifier string) (bool, error)
}

// RedisLimiter implements the Limiter interface using Redis and Lua for atomicity.
type RedisLimiter struct {
	client     *redis.Client
	luaScript  *redis.Script
	capacity   int64
	refillRate float64
}

// NewRedisLimiter initializes the distributed rate limiter
func NewRedisLimiter(redisAddr string, capacity int64, refillRate float64) (*RedisLimiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Quick ping to ensure Redis is running
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// The Lua script guarantees atomicity (no race conditions)
	const luaScript = `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local bucket = redis.call("HMGET", key, "tokens", "last_refill")
		local tokens = capacity
		local last_refill = now

		if bucket[1] then
			tokens = tonumber(bucket[1])
			last_refill = tonumber(bucket[2])
		end

		local elapsed = (now - last_refill) / 1000.0
		local tokens_to_add = elapsed * refill_rate
		tokens = math.min(capacity, tokens + tokens_to_add)

		local allowed = 0
		if tokens >= 1 then
			tokens = tokens - 1
			allowed = 1
		end

		-- EXPIRE prevents memory leaks in Redis!
		local ttl = math.ceil(capacity / refill_rate) + 10
		redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
		redis.call("EXPIRE", key, ttl)

		return allowed
	`

	return &RedisLimiter{
		client:     client,
		luaScript:  redis.NewScript(luaScript),
		capacity:   capacity,
		refillRate: refillRate,
	}, nil
}

// Allow implements the Limiter interface
func (rl *RedisLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	now := time.Now().UnixMilli()
	key := fmt.Sprintf("rate_limit:%s", identifier)

	// Run the Lua script atomically
	result, err := rl.luaScript.Run(ctx, rl.client, []string{key}, rl.capacity, rl.refillRate, now).Result()
	if err != nil {
		return false, fmt.Errorf("redis rate limit check failed: %w", err)
	}

	// The script returns 1 (allowed) or 0 (denied)
	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected redis response")
	}

	return allowed == 1, nil
}
