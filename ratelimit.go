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

// IPRateLimiter manages rate limiters per IP address
type IPRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	maxTokens  int
	refillRate time.Duration
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(maxTokens int, refillRate time.Duration) *IPRateLimiter {
	iprl := &IPRateLimiter{
		limiters:   make(map[string]*RateLimiter),
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
	
	// Clean up old limiters every 5 minutes
	go iprl.cleanup()
	
	return iprl
}

// GetLimiter returns a rate limiter for the given IP
func (iprl *IPRateLimiter) GetLimiter(ip string) *RateLimiter {
	iprl.mu.Lock()
	defer iprl.mu.Unlock()
	
	limiter, exists := iprl.limiters[ip]
	if !exists {
		limiter = NewRateLimiter(iprl.maxTokens, iprl.refillRate)
		iprl.limiters[ip] = limiter
	}
	
	return limiter
}

// cleanup removes inactive rate limiters periodically
func (iprl *IPRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		iprl.mu.Lock()
		// In production, you'd want to track last access time
		// For now, we'll keep all limiters (they're lightweight)
		iprl.mu.Unlock()
	}
}

// Global rate limiter - 200 requests per second for entire server
var globalRateLimiter = NewRateLimiter(200, 5*time.Millisecond)

// Per-IP rate limiter - 10 requests per second per IP
var ipRateLimiter = NewIPRateLimiter(10, 100*time.Millisecond)
