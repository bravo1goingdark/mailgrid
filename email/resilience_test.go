package email

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_Basic(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// Initially closed
	if cb.state != Closed {
		t.Error("Circuit breaker should start closed")
	}

	// Test successful calls
	ctx := context.Background()
	err := cb.Call(ctx, func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Successful call should not return error: %v", err)
	}

	stats := cb.GetState()
	if stats.Successes != 1 {
		t.Error("Expected 1 success")
	}
}

func TestCircuitBreaker_Failure(t *testing.T) {
	cb := NewCircuitBreaker(2, 100*time.Millisecond)
	ctx := context.Background()

	testErr := errors.New("network error")

	// First failure - should still be closed
	err := cb.Call(ctx, func() error {
		return testErr
	})
	if err != testErr {
		t.Error("Should return the original error")
	}
	if cb.state != Closed {
		t.Error("Should still be closed after 1 failure")
	}

	// Second failure - should open (threshold is 2)
	cb.Call(ctx, func() error {
		return testErr
	})
	if cb.state != Open {
		t.Error("Should be open after 2 failures")
	}

	// Subsequent call should fail immediately
	err = cb.Call(ctx, func() error {
		return nil // This shouldn't be called
	})
	if err != ErrCircuitBreakerOpen {
		t.Error("Should return circuit breaker open error")
	}
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	ctx := context.Background()

	// Cause it to open
	cb.Call(ctx, func() error {
		return errors.New("network error")
	})
	cb.Call(ctx, func() error {
		return errors.New("network error")
	})

	if cb.state != Open {
		t.Error("Should be open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Should transition to half-open on next call
	err := cb.Call(ctx, func() error {
		return nil // Success
	})
	if err != nil {
		t.Error("Should succeed and close circuit")
	}
	if cb.state != Closed {
		t.Error("Should be closed after successful half-open call")
	}
}

func TestErrorClassifier_Classification(t *testing.T) {
	classifier := NewErrorClassifier()

	tests := []struct {
		err      error
		expected ErrorType
	}{
		{errors.New("connection refused"), NetworkError},
		{errors.New("authentication failed"), AuthError},
		{errors.New("quota exceeded"), QuotaError},
		{errors.New("temporary failure"), TemporaryError},
		{errors.New("invalid recipient"), PermanentError},
		{errors.New("unknown error"), UnknownError},
	}

	for _, test := range tests {
		result := classifier.ClassifyError(test.err)
		if result != test.expected {
			t.Errorf("Expected %v for error %q, got %v", test.expected, test.err.Error(), result)
		}
	}
}

func TestRetryPolicy_Basic(t *testing.T) {
	policy := DefaultRetryPolicy()
	classifier := NewErrorClassifier()
	ctx := context.Background()

	attempts := 0
	err := policy.Retry(ctx, classifier, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Should succeed after retries: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryPolicy_NonRetryableError(t *testing.T) {
	policy := DefaultRetryPolicy()
	classifier := NewErrorClassifier()
	ctx := context.Background()

	attempts := 0
	err := policy.Retry(ctx, classifier, func() error {
		attempts++
		return errors.New("permanent failure") // Non-retryable
	})

	if err == nil {
		t.Error("Should fail with permanent error")
	}
	if attempts != 1 {
		t.Errorf("Should only attempt once for permanent error, got %d", attempts)
	}
}

func TestRetryPolicy_MaxRetries(t *testing.T) {
	policy := &RetryPolicy{
		MaxRetries:    2,
		BaseDelay:     1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
		RetryableErrors: map[ErrorType]bool{
			NetworkError: true,
		},
	}
	classifier := NewErrorClassifier()
	ctx := context.Background()

	attempts := 0
	err := policy.Retry(ctx, classifier, func() error {
		attempts++
		return errors.New("connection refused") // Retryable
	})

	if err == nil {
		t.Error("Should fail after max retries")
	}
	if attempts != 3 { // Initial + 2 retries
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestResilienceManager_Integration(t *testing.T) {
	rm := NewResilienceManager(2, 100*time.Millisecond, nil)
	ctx := context.Background()

	attempts := 0
	err := rm.Execute(ctx, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary failure")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Should succeed with resilience manager: %v", err)
	}

	stats := rm.GetStats()
	if stats.CircuitBreaker.Successes == 0 {
		t.Error("Should record success")
	}
}

func TestResilienceManager_CircuitBreakerIntegration(t *testing.T) {
	rm := NewResilienceManager(1, 50*time.Millisecond, nil)
	ctx := context.Background()

	// Cause circuit breaker to open
	rm.Execute(ctx, func() error {
		return errors.New("network error")
	})
	rm.Execute(ctx, func() error {
		return errors.New("network error")
	})

	// Should fail immediately due to open circuit
	err := rm.Execute(ctx, func() error {
		return nil
	})
	if err != ErrCircuitBreakerOpen {
		t.Error("Should fail due to open circuit breaker")
	}
}