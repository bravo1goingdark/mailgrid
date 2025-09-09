package ratelimit

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting for email sending
type RateLimiter struct {
	limiter *rate.Limiter
	mu      sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
// emailsPerSecond: maximum emails per second (0 = unlimited)
// burstSize: maximum burst size
func NewRateLimiter(emailsPerSecond int, burstSize int) *RateLimiter {
	if emailsPerSecond <= 0 {
		// Unlimited rate
		return &RateLimiter{
			limiter: rate.NewLimiter(rate.Inf, 0),
		}
	}
	
	if burstSize <= 0 {
		burstSize = emailsPerSecond // Default burst equals rate
	}
	
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(emailsPerSecond), burstSize),
	}
}

// Wait blocks until the rate limiter allows the operation
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.RLock()
	limiter := rl.limiter
	rl.mu.RUnlock()
	
	return limiter.Wait(ctx)
}

// Allow returns true if the operation is allowed immediately
func (rl *RateLimiter) Allow() bool {
	rl.mu.RLock()
	limiter := rl.limiter
	rl.mu.RUnlock()
	
	return limiter.Allow()
}

// SetRate updates the rate limiting configuration
func (rl *RateLimiter) SetRate(emailsPerSecond int, burstSize int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	if emailsPerSecond <= 0 {
		rl.limiter.SetLimit(rate.Inf)
		rl.limiter.SetBurst(0)
		return
	}
	
	if burstSize <= 0 {
		burstSize = emailsPerSecond
	}
	
	rl.limiter.SetLimit(rate.Limit(emailsPerSecond))
	rl.limiter.SetBurst(burstSize)
}

// GetCurrentRate returns the current rate limit settings
func (rl *RateLimiter) GetCurrentRate() (limit float64, burst int) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	return float64(rl.limiter.Limit()), rl.limiter.Burst()
}
