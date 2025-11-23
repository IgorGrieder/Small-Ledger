package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Currency string    `json:"currency"`
}

type TransactionRequest struct {
	From           uuid.UUID `json:"from"`
	To             uuid.UUID `json:"to"`
	Currency       string    `json:"currency"`
	Amount         int64     `json:"amount"`
	IdempotencyKey uuid.UUID `json:"idempotency_key"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 1 * time.Minute,
		},
	}
}

func (c *Client) GetAccounts() ([]Account, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/accounts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var accounts []Account
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (c *Client) SendTransaction(req TransactionRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/transaction", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
