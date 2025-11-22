package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type HTTPResponse struct {
	URL      string
	Response *http.Response
	Error    error
}

// ConcurrentRequest defines the parameters for a single concurrent call
type ConcurrentRequest struct {
	URL         string
	Method      string
	QueryParams map[string]string
	Headers     map[string]string
	Body        any
}

type Client struct {
	client *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

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

func (c *Client) PostWithJson(ctx context.Context, url string, body any, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.client.Do(req)
}

func (c *Client) FetchConcurrent(ctx context.Context, requests []ConcurrentRequest) <-chan HTTPResponse {
	results := make(chan HTTPResponse, len(requests))
	var wg sync.WaitGroup

	for _, req := range requests {
		wg.Add(1)
		go func(r ConcurrentRequest) {
			defer wg.Done()
			var resp *http.Response
			var err error

			switch r.Method {
			case http.MethodGet:
				resp, err = c.Get(ctx, r.URL, r.QueryParams, r.Headers)
			case http.MethodPost:
				resp, err = c.PostWithJson(ctx, r.URL, r.Body, r.Headers)
			default:
				resp, err = nil, http.ErrNotSupported
			}

			results <- HTTPResponse{
				URL:      r.URL,
				Response: resp,
				Error:    err,
			}
		}(req)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
