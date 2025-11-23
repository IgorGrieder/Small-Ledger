package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is a wrapper around http.Client that provides convenience methods
// for making HTTP requests with context and timeouts.
type Client struct {
	client *http.Client
}

// NewClient creates a new Client with the specified timeout.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get performs a GET request to the specified URL with query parameters.
func (c *Client) Get(ctx context.Context, baseURL string, queryParams map[string]string, headers map[string]string) (*http.Response, error) {
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

	return c.client.Do(req)
}

// Post performs a POST request to the specified URL with the given body marshaled to JSON.
func (c *Client) Post(ctx context.Context, url string, body any, headers map[string]string) (*http.Response, error) {
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

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.client.Do(req)
}
