package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected X-Test-Header to be test-value, got %s", r.Header.Get("X-Test-Header"))
		}
		if r.URL.Query().Get("q") != "golang" {
			t.Errorf("Expected query param q=golang, got %s", r.URL.Query().Get("q"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	headers := map[string]string{"X-Test-Header": "test-value"}
	queryParams := map[string]string{"q": "golang"}

	resp, err := client.Get(context.Background(), server.URL, queryParams, headers)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}

func TestClient_Post(t *testing.T) {
	type RequestBody struct {
		Name string `json:"name"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var reqBody RequestBody
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Errorf("Failed to unmarshal body: %v", err)
		}
		if reqBody.Name != "test-name" {
			t.Errorf("Expected name test-name, got %s", reqBody.Name)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	headers := map[string]string{"X-Custom-Header": "custom-value"}
	body := RequestBody{Name: "test-name"}

	resp, err := client.Post(context.Background(), server.URL, body, headers)
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status Created, got %v", resp.Status)
	}
}

func TestClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(10 * time.Millisecond)

	_, err := client.Get(context.Background(), server.URL, nil, nil)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestClient_Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	// Override CB to avoid opening it during retries if thresholds are low
	client.cb = NewCircuitBreaker(5, 10*time.Second)

	resp, err := client.Get(context.Background(), server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Failed to make request with retries: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestClient_CircuitBreaker_Integration(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// CB opens after 2 failures
	client := NewClient(5 * time.Second)
	client.cb = NewCircuitBreaker(2, 100*time.Millisecond)

	// 1st request (fails, retries 3 times -> 4 attempts total? No, max retries is 3, so 1 initial + 3 retries = 4 attempts)
	// Wait, my implementation does 0 to MAX_RETRIES loop.
	// i=0 (attempt 1), i=1 (attempt 2), i=2 (attempt 3), i=3 (attempt 4).
	// So 4 attempts per call.
	// CB counts failures.
	// If max failures is 2.
	// Call 1:
	// Attempt 1: 500 -> CB failure 1
	// Attempt 2: 500 -> CB failure 2 -> Open?
	// Wait, OnFailure is called AFTER the loop finishes or if CheckBeforeRequest fails?
	// In my implementation:
	// CheckBeforeRequest is called ONCE at the start.
	// Then the loop runs.
	// If loop fails (all retries), OnFailure is called ONCE.
	// So 1 Call = 1 Failure count in CB.

	// Call 1
	_, err := client.Get(context.Background(), server.URL, nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}

	// Call 2
	_, err = client.Get(context.Background(), server.URL, nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}

	// CB should be open now (2 failures)

	// Call 3 - Should be blocked immediately
	_, err = client.Get(context.Background(), server.URL, nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}
	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}
