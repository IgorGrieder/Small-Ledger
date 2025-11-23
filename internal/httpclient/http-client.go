package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// Client is a wrapper around http.Client that provides convenience methods
// for making HTTP requests with context and timeouts.
type Client struct {
	client *http.Client
	cb     *CircuitBreaker
}

// NewClient creates a new Client with the specified timeout.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
		},
		cb: NewCircuitBreaker(5, 10*time.Second),
	}
}

// Get performs a GET request to the specified URL with query parameters.
func (c *Client) Get(ctx context.Context, baseURL string, queryParams map[string]string, headers map[string]string) (*http.Response, error) {
	return c.attemptRequestWithRetry(ctx, func() (*http.Request, error) {
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}

		q := u.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req, nil
	})
}

// Post performs a POST request to the specified URL with the given body marshaled to JSON.
func (c *Client) Post(ctx context.Context, url string, body any, headers map[string]string) (*http.Response, error) {
	return c.attemptRequestWithRetry(ctx, func() (*http.Request, error) {
		var bodyReader io.Reader
		if body != nil {
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewBuffer(jsonData)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
		if err != nil {
			return nil, err
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req, nil
	})
}

func (c *Client) attemptRequestWithRetry(ctx context.Context, reqFactory func() (*http.Request, error)) (*http.Response, error) {
	if err := c.cb.CheckBeforeRequest(); err != nil {
		slog.Error("Request blocked by circuit breaker", slog.String("error", err.Error()))
		return nil, err
	}

	const MAX_RETRIES = 3
	const BASE_DELAY = 100 * time.Millisecond
	const MAX_JITTER_MS = 100

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var lastErr error
	var response *http.Response

	for i := 0; i <= MAX_RETRIES; i++ {
		req, err := reqFactory()
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		response, err = c.client.Do(req)
		lastErr = err

		if err == nil && response.StatusCode < 500 {
			c.cb.OnSuccess()
			return response, nil
		}

		if i == MAX_RETRIES {
			break
		}

		// Calculate backoff and jitter
		backoff := BASE_DELAY * time.Duration(math.Pow(2, float64(i)))
		jitter := time.Duration(r.Intn(MAX_JITTER_MS)) * time.Millisecond
		sleepDuration := backoff + jitter

		errMsg := "server error"
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = response.Status
			response.Body.Close() // Close body before retrying
		}

		slog.Warn("Request failed, retrying",
			slog.Int("attempt", i+1),
			slog.String("sleep_duration", sleepDuration.String()),
			slog.String("error", errMsg),
		)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleepDuration):
		}
	}

	c.cb.OnFailure()

	if lastErr != nil {
		return nil, fmt.Errorf("all retries failed, last network error: %w", lastErr)
	}

	return nil, fmt.Errorf("all retries failed, last status: %s", response.Status)
}
