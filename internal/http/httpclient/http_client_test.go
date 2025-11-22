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

	resp, err := client.PostWithJson(context.Background(), server.URL, body, headers)
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
