package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPipeline(t *testing.T) {
	p := NewPipeline(10)

	assert.NotNil(t, p)
	assert.NotNil(t, p.commands)
	assert.NotNil(t, p.results)
	assert.Equal(t, 0, len(p.commands))
	assert.Equal(t, 10, cap(p.results))
}

func TestPipeline_QueueCommand(t *testing.T) {
	p := NewPipeline(5)

	p.QueueCommand("MAIL FROM:<%s>", "test@example.com")
	p.QueueCommand("RCPT TO:<%s>", "recipient@example.com")

	assert.Equal(t, 2, len(p.commands))
	assert.Equal(t, "MAIL FROM:<test@example.com>", p.commands[0])
	assert.Equal(t, "RCPT TO:<recipient@example.com>", p.commands[1])
}

func TestNewSMTPPipeline(t *testing.T) {
	pipeline := NewSMTPPipeline(nil, 50, 100*time.Millisecond)

	assert.NotNil(t, pipeline)
	assert.Equal(t, 50, pipeline.maxPipeline)
	assert.Equal(t, 100*time.Millisecond, pipeline.flushInterval)
}

func TestNewSMTPPipelineDefaultValues(t *testing.T) {
	pipeline := NewSMTPPipeline(nil, 0, 0)

	assert.NotNil(t, pipeline)
	assert.Equal(t, 100, pipeline.maxPipeline)
	assert.Equal(t, 100*time.Millisecond, pipeline.flushInterval)
}

func TestNewErrorClassifier(t *testing.T) {
	classifier := NewErrorClassifier()

	assert.NotNil(t, classifier)
	assert.NotNil(t, classifier.patterns)
	assert.Greater(t, len(classifier.patterns), 0)
}

func TestErrorClassifier_ClassifyError(t *testing.T) {
	classifier := NewErrorClassifier()

	tests := []struct {
		name        string
		err         error
		expectedErr ErrorType
	}{
		{
			name:        "network error - connection refused",
			err:         &testError{"connection refused"},
			expectedErr: NetworkError,
		},
		{
			name:        "network error - timeout",
			err:         &testError{"timeout occurred"},
			expectedErr: NetworkError,
		},
		{
			name:        "auth error",
			err:         &testError{"authentication failed"},
			expectedErr: AuthError,
		},
		{
			name:        "quota error",
			err:         &testError{"quota exceeded"},
			expectedErr: QuotaError,
		},
		{
			name:        "rate limit error",
			err:         &testError{"rate limit exceeded"},
			expectedErr: QuotaError,
		},
		{
			name:        "temporary error",
			err:         &testError{"temporary failure"},
			expectedErr: TemporaryError,
		},
		{
			name:        "permanent error",
			err:         &testError{"permanent failure"},
			expectedErr: PermanentError,
		},
		{
			name:        "invalid recipient",
			err:         &testError{"invalid recipient"},
			expectedErr: PermanentError,
		},
		{
			name:        "unknown error",
			err:         &testError{"some random error"},
			expectedErr: UnknownError,
		},
		{
			name:        "nil error",
			err:         nil,
			expectedErr: UnknownError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyError(tt.err)
			assert.Equal(t, tt.expectedErr, result)
		})
	}
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(10, 60*time.Second)

	assert.NotNil(t, cb)
	assert.Equal(t, int64(10), cb.maxFailures)
	assert.Equal(t, 60*time.Second, cb.timeout)
	assert.Equal(t, Closed, cb.state)
}

func TestNewCircuitBreakerDefaultValues(t *testing.T) {
	cb := NewCircuitBreaker(0, 0)

	assert.NotNil(t, cb)
	assert.Equal(t, int64(5), cb.maxFailures)
	assert.Equal(t, 60*time.Second, cb.timeout)
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	cb := NewCircuitBreaker(5, 60*time.Second)

	// Record some failures first
	cb.recordFailure(&testError{"error 1"})
	cb.recordFailure(&testError{"error 2"})

	stats := cb.GetState()
	assert.Equal(t, int64(2), stats.Failures)

	// Record success
	cb.recordSuccess()

	stats = cb.GetState()
	assert.Equal(t, int64(1), stats.Failures)
	assert.Equal(t, int64(1), stats.Successes)
}

func TestCircuitBreaker_RecordFailure(t *testing.T) {
	cb := NewCircuitBreaker(3, 60*time.Second)

	cb.recordFailure(&testError{"error 1"})

	stats := cb.GetState()
	assert.Equal(t, int64(1), stats.Failures)
	assert.Equal(t, Closed, stats.State)

	// Record more failures to trip the circuit
	cb.recordFailure(&testError{"error 2"})
	cb.recordFailure(&testError{"error 3"})

	stats = cb.GetState()
	assert.Equal(t, int64(3), stats.Failures)
	assert.Equal(t, Open, stats.State)
}

func TestCircuitBreaker_TransitionToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	// Trip the circuit
	cb.recordFailure(&testError{"error 1"})
	cb.recordFailure(&testError{"error 2"})

	stats := cb.GetState()
	assert.Equal(t, Open, stats.State)

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Check if request is allowed (should transition to HalfOpen)
	allowed := cb.allowRequest()
	assert.True(t, allowed)

	stats = cb.GetState()
	assert.Equal(t, HalfOpen, stats.State)
}

func TestCircuitBreaker_GetState(t *testing.T) {
	cb := NewCircuitBreaker(5, 60*time.Second)

	cb.recordFailure(&testError{"network error"})

	stats := cb.GetState()

	assert.Equal(t, Closed, stats.State)
	assert.Equal(t, int64(1), stats.Failures)
	assert.Equal(t, int64(0), stats.Successes)
	assert.NotZero(t, stats.LastFailTime)
}

func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy()

	assert.NotNil(t, policy)
	assert.Equal(t, 3, policy.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, policy.BaseDelay)
	assert.Equal(t, 10*time.Second, policy.MaxDelay)
	assert.Equal(t, 2.0, policy.BackoffFactor)
	assert.True(t, policy.RetryableErrors[NetworkError])
	assert.True(t, policy.RetryableErrors[TemporaryError])
	assert.False(t, policy.RetryableErrors[PermanentError])
}

func TestNewResilienceManager(t *testing.T) {
	policy := DefaultRetryPolicy()
	rm := NewResilienceManager(5, 60*time.Second, policy)

	assert.NotNil(t, rm)
	assert.NotNil(t, rm.circuitBreaker)
	assert.NotNil(t, rm.retryPolicy)
	assert.NotNil(t, rm.classifier)
}

func TestNewResilienceManagerNilPolicy(t *testing.T) {
	rm := NewResilienceManager(5, 60*time.Second, nil)

	assert.NotNil(t, rm)
	assert.NotNil(t, rm.retryPolicy)
}

func TestResilienceManager_GetStats(t *testing.T) {
	rm := NewResilienceManager(5, 60*time.Second, nil)

	stats := rm.GetStats()

	assert.Equal(t, Closed, stats.CircuitBreaker.State)
}
