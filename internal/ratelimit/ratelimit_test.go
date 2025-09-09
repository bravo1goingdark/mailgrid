package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(10, 5)

	count := 0
	for i := 0; i < 5; i++ {
		if rl.Allow() {
			count++
		}
	}

	if count != 5 {
		t.Errorf("Expected 5 immediate allows within burst, got %d", count)
	}
}

func TestRateLimiterWait(t *testing.T) {
	rl := NewRateLimiter(2, 2) // 2 per second
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	for i := 0; i < 4; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}
	duration := time.Since(start)

	// 4 operations at 2/s should take roughly >= 1 second (with some tolerance)
	if duration < 900*time.Millisecond {
		t.Errorf("Expected duration >= 0.9s, got %v", duration)
	}
}

func TestRateLimiterUnlimited(t *testing.T) {
	rl := NewRateLimiter(0, 0) // unlimited
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	for i := 0; i < 100; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Logf("Unlimited rate limiter took longer than expected: %v", time.Since(start))
	}
}

