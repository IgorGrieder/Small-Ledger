package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type transactionRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

func transactionHandler(w http.ResponseWriter, r *http.Request) {
	var request transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "error decoding request json", http.StatusBadRequest)
		slog.Error("error decoding json")
		return
	}

}
