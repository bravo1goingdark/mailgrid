package email

import (
	"context"
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"
)

// CircuitBreakerState represents the current state of a circuit breaker
type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	Open
	HalfOpen
)

// ErrorClassifier defines different types of errors
type ErrorType int

const (
	UnknownError ErrorType = iota
	NetworkError
	AuthError
	QuotaError
	TemporaryError
	PermanentError
)

// ErrorClassifier classifies errors for circuit breaker decisions
type ErrorClassifier struct {
	patterns map[string]ErrorType
}

// NewErrorClassifier creates a new error classifier
func NewErrorClassifier() *ErrorClassifier {
	return &ErrorClassifier{
		patterns: map[string]ErrorType{
			"connection refused":  NetworkError,
			"timeout":             NetworkError,
			"authentication":      AuthError,
			"quota":               QuotaError,
			"rate limit":          QuotaError,
			"temporary":           TemporaryError,
			"mailbox unavailable": TemporaryError,
			"invalid recipient":   PermanentError,
			"permanent failure":   PermanentError,
		},
	}
}

// ClassifyError determines the type of error
func (c *ErrorClassifier) ClassifyError(err error) ErrorType {
	if err == nil {
		return UnknownError
	}

	errStr := strings.ToLower(err.Error())
	for pattern, errorType := range c.patterns {
		if strings.Contains(errStr, pattern) {
			return errorType
		}
	}
	return UnknownError
}

// CircuitBreaker implements circuit breaker pattern for email sending
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	maxFailures  int64
	timeout      time.Duration
	resetTimeout time.Duration

	// State
	state        CircuitBreakerState
	failures     int64
	successes    int64
	lastFailTime time.Time
	nextAttempt  time.Time

	// Error tracking
	classifier      *ErrorClassifier
	errorCounts     map[ErrorType]int64
	recentErrors    []error
	maxRecentErrors int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int64, timeout time.Duration) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &CircuitBreaker{
		maxFailures:     maxFailures,
		timeout:         timeout,
		resetTimeout:    timeout * 2,
		state:           Closed,
		classifier:      NewErrorClassifier(),
		errorCounts:     make(map[ErrorType]int64),
		recentErrors:    make([]error, 0, 100),
		maxRecentErrors: 100,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
	// Check if circuit is open
	if !cb.allowRequest() {
		return ErrCircuitBreakerOpen
	}

	// Execute the function
	err := fn()

	// Record the result
	if err != nil {
		cb.recordFailure(err)
		return err
	}

	cb.recordSuccess()
	return nil
}

// allowRequest determines if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case Closed:
		return true
	case Open:
		if now.After(cb.nextAttempt) {
			cb.state = HalfOpen
			return true
		}
		return false
	case HalfOpen:
		return true
	default:
		return false
	}
}

// recordSuccess records a successful operation
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successes++

	switch cb.state {
	case HalfOpen:
		// Reset the circuit breaker
		cb.state = Closed
		cb.failures = 0
	case Closed:
		// Decay failure count on success
		if cb.failures > 0 {
			cb.failures--
		}
	}
}

// recordFailure records a failed operation
func (cb *CircuitBreaker) recordFailure(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	errorType := cb.classifier.ClassifyError(err)
	cb.errorCounts[errorType]++

	// Add to recent errors
	cb.recentErrors = append(cb.recentErrors, err)
	if len(cb.recentErrors) > cb.maxRecentErrors {
		cb.recentErrors = cb.recentErrors[1:]
	}

	// Count all error types towards circuit breaking for now
	// In production, you might want to be more selective

	cb.failures++
	cb.lastFailTime = time.Now()

	// Check if we should trip the circuit
	if cb.state == Closed && cb.failures >= cb.maxFailures {
		cb.state = Open
		cb.nextAttempt = time.Now().Add(cb.timeout)
	} else if cb.state == HalfOpen {
		// Failed in half-open, go back to open
		cb.state = Open
		cb.nextAttempt = time.Now().Add(cb.resetTimeout)
	}
}

// GetState returns current circuit breaker state and metrics
func (cb *CircuitBreaker) GetState() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:        cb.state,
		Failures:     cb.failures,
		Successes:    cb.successes,
		LastFailTime: cb.lastFailTime,
		NextAttempt:  cb.nextAttempt,
		ErrorCounts:  cb.copyErrorCounts(),
		RecentErrors: cb.copyRecentErrors(),
	}
}

func (cb *CircuitBreaker) copyErrorCounts() map[ErrorType]int64 {
	result := make(map[ErrorType]int64)
	for k, v := range cb.errorCounts {
		result[k] = v
	}
	return result
}

func (cb *CircuitBreaker) copyRecentErrors() []string {
	result := make([]string, len(cb.recentErrors))
	for i, err := range cb.recentErrors {
		result[i] = err.Error()
	}
	return result
}

// CircuitBreakerStats provides metrics about circuit breaker state
type CircuitBreakerStats struct {
	State        CircuitBreakerState
	Failures     int64
	Successes    int64
	LastFailTime time.Time
	NextAttempt  time.Time
	ErrorCounts  map[ErrorType]int64
	RecentErrors []string
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors map[ErrorType]bool
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: map[ErrorType]bool{
			NetworkError:   true,
			TemporaryError: true,
			QuotaError:     true,
			UnknownError:   false,
			AuthError:      false,
			PermanentError: false,
		},
	}
}

// Retry executes a function with retry logic
func (rp *RetryPolicy) Retry(ctx context.Context, classifier *ErrorClassifier, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= rp.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := time.Duration(float64(rp.BaseDelay) * math.Pow(rp.BackoffFactor, float64(attempt-1)))
			if delay > rp.MaxDelay {
				delay = rp.MaxDelay
			}

			// Add jitter to prevent thundering herd
			jitterMax := int64(delay) / 4
			if jitterMax <= 0 {
				jitterMax = 1
			}
			jitterNs, _ := rand.Int(rand.Reader, big.NewInt(jitterMax))
			jitter := time.Duration(jitterNs.Int64())
			delay += jitter

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		errorType := classifier.ClassifyError(err)

		// Check if this error type is retryable
		if retryable, ok := rp.RetryableErrors[errorType]; !ok || !retryable {
			return err
		}
	}

	return lastErr
}

// ResilienceManager combines circuit breaker and retry logic
type ResilienceManager struct {
	circuitBreaker *CircuitBreaker
	retryPolicy    *RetryPolicy
	classifier     *ErrorClassifier
}

// NewResilienceManager creates a new resilience manager
func NewResilienceManager(maxFailures int64, timeout time.Duration, retryPolicy *RetryPolicy) *ResilienceManager {
	if retryPolicy == nil {
		retryPolicy = DefaultRetryPolicy()
	}

	return &ResilienceManager{
		circuitBreaker: NewCircuitBreaker(maxFailures, timeout),
		retryPolicy:    retryPolicy,
		classifier:     NewErrorClassifier(),
	}
}

// Execute runs a function with both circuit breaker and retry protection
func (rm *ResilienceManager) Execute(ctx context.Context, fn func() error) error {
	return rm.circuitBreaker.Call(ctx, func() error {
		return rm.retryPolicy.Retry(ctx, rm.classifier, fn)
	})
}

// GetStats returns current resilience manager statistics
func (rm *ResilienceManager) GetStats() ResilienceStats {
	return ResilienceStats{
		CircuitBreaker: rm.circuitBreaker.GetState(),
	}
}

// ResilienceStats provides overall resilience statistics
type ResilienceStats struct {
	CircuitBreaker CircuitBreakerStats
}

var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)
