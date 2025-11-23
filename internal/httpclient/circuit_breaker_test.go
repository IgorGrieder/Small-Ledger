package httpclient

import (
	"testing"
	"time"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	// Create a circuit breaker with 2 max failures and 100ms timeout
	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	// Initial state should be Closed
	if err := cb.CheckBeforeRequest(); err != nil {
		t.Errorf("Expected circuit to be closed, got error: %v", err)
	}

	// 1st failure
	cb.OnFailure()
	if cb.state != StateClosed {
		t.Errorf("Expected state to be Closed after 1 failure, got %v", cb.state)
	}

	// 2nd failure -> Should trip to Open
	cb.OnFailure()
	if cb.state != StateOpen {
		t.Errorf("Expected state to be Open after 2 failures, got %v", cb.state)
	}

	// Request should be blocked
	if err := cb.CheckBeforeRequest(); err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Next request should transition to Half-Open
	if err := cb.CheckBeforeRequest(); err != nil {
		t.Errorf("Expected to be allowed (Half-Open), got error: %v", err)
	}
	if cb.state != StateHalfOpen {
		t.Errorf("Expected state to be Half-Open, got %v", cb.state)
	}

	// Another concurrent request should be blocked while in Half-Open
	if err := cb.CheckBeforeRequest(); err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen for concurrent request in Half-Open, got %v", err)
	}

	// Success -> Should transition back to Closed
	cb.OnSuccess()
	if cb.state != StateClosed {
		t.Errorf("Expected state to be Closed after success, got %v", cb.state)
	}
	if cb.failures != 0 {
		t.Errorf("Expected failures to be reset to 0, got %d", cb.failures)
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(1, 100*time.Millisecond)

	// Trip to Open
	cb.OnFailure()
	if cb.state != StateOpen {
		t.Fatalf("Expected state Open")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Transition to Half-Open
	cb.CheckBeforeRequest()
	if cb.state != StateHalfOpen {
		t.Fatalf("Expected state Half-Open")
	}

	// Failure in Half-Open -> Should go back to Open
	cb.OnFailure()
	if cb.state != StateOpen {
		t.Errorf("Expected state to be Open after failure in Half-Open, got %v", cb.state)
	}
}
