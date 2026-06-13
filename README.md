# Distributed Rate-Limited HTTP Server in Go

A production-grade HTTP API server built in Go, featuring a distributed, atomic Token Bucket rate limiter powered by Redis and Lua. 

## Key Features
* **Distributed Rate Limiting:** Uses Redis and atomic Lua scripts to guarantee zero race conditions across multiple server instances.
* **Resilience & Context Timeouts:** Strict 50ms context deadlines on Redis calls prevent network hangs from cascading into server outages.
* **Graceful Shutdown:** Listens for OS signals (SIGINT, SIGTERM) to cleanly close connections and finish in-flight requests.
* **Thread-Safe Data Store:** In-memory User store protected by sync.RWMutex for safe concurrent reads/writes.

## Prerequisites
* Go 1.22+ (https://go.dev/doc/install)
* Redis (Running locally on port 6379)

## How to Run

1. Start Redis (if not already running):
   sudo service redis-server start
   Verify it's running:
   redis-cli ping  # Should output: PONG

2. Start the Go Server:
   cd cmd/server
   go run .
   
   You should see: "Server started on port 8080."

## How to Test

To test the rate limiter, we will create a user and immediately spam the GET endpoint to exhaust the token bucket. 

Run this chained command in a second terminal window:

curl -s -X POST -H "Content-Type: application/json" -d '{"name":"TestUser"}' http://localhost:8080/users > /dev/null && \
for i in {1..10}; do curl -s -o /dev/null -w "Request $i: %{http_code}\n" http://localhost:8080/users/1; done

## What to Expect

Because the server is configured with a capacity of 10 tokens and a refill rate of 1 token/second, the first request (the POST) consumes a token. The subsequent loop will rapidly consume the remaining tokens.

You should see the first 9 requests succeed, and the 10th request get blocked:

Request 1: 200
Request 2: 200
Request 3: 200
Request 4: 200
Request 5: 200
Request 6: 200
Request 7: 200
Request 8: 200
Request 9: 200
Request 10: 429  <-- Rate limit exceeded!

*Note: If you wait 5 seconds and run the loop again, you will see the bucket has refilled, allowing 5 more successful requests.*

## Architecture
* cmd/server: Application entry point and graceful shutdown logic.
* internal/handlers: HTTP request handlers.
* internal/middleware: Rate limiting middleware with context timeouts and fail-open/fail-closed strategies.
* internal/rate-limiter: Redis client and atomic Lua script implementation.
* internal/store: Thread-safe in-memory data store.