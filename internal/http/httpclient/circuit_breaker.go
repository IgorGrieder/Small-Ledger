package httpclient

import (
	"errors"
	"log/slog"
	"sync"
	"time"
)

// Define the states for the circuit breaker
type State int

const (
	StateClosed State = iota + 1
	StateOpen
	StateHalfOpen
)

// Sentinel error
var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu          sync.Mutex
	state       State
	failures    int
	maxFailures int
	openSince   time.Time
	openTimeout time.Duration
}

func NewCircuitBreaker(maxFailures int, openTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		openTimeout: openTimeout,
	}
}

func (cb *CircuitBreaker) CheckBeforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Always allowed in a closed state
		return nil

	case StateOpen:
		// Check if the open timeout has elapsed
		if time.Since(cb.openSince) > cb.openTimeout {
			// Timeout exceeded -> Half-Open
			slog.Warn("Circuit Breaker: Open -> Half-Open")
			cb.state = StateHalfOpen
			return nil // Allow one test request to go through
		}

		// Still open
		return ErrCircuitOpen

	case StateHalfOpen:
		// The circuit is already in a Half-Open state
		// a test request is in flight. Reject all other concurrent requests
		return ErrCircuitOpen
	}
	return nil
}

// OnSuccess notifies the breaker of a successful call
func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		// Test request succeeded -> close circuit
		slog.Info("Circuit Breaker: Half-Open -> Closed")
		cb.state = StateClosed
		cb.failures = 0

	case StateClosed:
		// Reset consecutive failures
		cb.failures = 0
	}
}

// OnFailure notifies the breaker of a failed call
func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		// The test request failed -> go into open state again
		slog.Error("Circuit Breaker: Half-Open -> Open (test failed)")
		cb.state = StateOpen
		cb.openSince = time.Now() // Reset the open timer

	case StateClosed:
		cb.failures++
		slog.Warn("Circuit Breaker: Failure recorded", "count", cb.failures)

		// Check if we've reached the threshold
		if cb.failures >= cb.maxFailures {
			slog.Error("Circuit Breaker: Closed -> Open (threshold reached)")
			cb.state = StateOpen
			cb.openSince = time.Now()
		}
	}
}
