package main

import (
	"sync"
	"time"
)

// RateLimiter controls the rate of requests
type RateLimiter struct {
	tokens     chan struct{}
	maxTokens  int
	refillRate time.Duration
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(chan struct{}, maxTokens),
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
	
	// Fill initial tokens
	for i := 0; i < maxTokens; i++ {
		rl.tokens <- struct{}{}
	}
	
	// Start refill goroutine
	go rl.refill()
	
	return rl
}

func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()
	
	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Token bucket is full
		}
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	<-rl.tokens
}

// Global rate limiter - 50 requests per second
var globalRateLimiter = NewRateLimiter(50, 20*time.Millisecond)
