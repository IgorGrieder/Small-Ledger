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

func (c *Client) FetchConcurrentUrls(ctx context.Context, wg *sync.WaitGroup, urls map[string]chan HTTPResponse, method string, headers map[string]map[string]string, queryParams map[string]map[string]string) {
	for targetURL, ch := range urls {
		wg.Add(1)

		go func(u string, outCh chan HTTPResponse) {
			defer wg.Done()
			defer close(outCh)

			var resp *http.Response
			var err error

			switch method {
			case http.MethodGet:
				resp, err = c.Get(ctx, u, queryParams[targetURL], headers[targetURL])
			case http.MethodPost:
				resp, err = c.Post(ctx, u, queryParams[targetURL], headers[targetURL])
			default:
				return
			}

			outCh <- HTTPResponse{
				URL:      u,
				Response: resp,
				Error:    err,
			}
		}(targetURL, ch)
	}
}
