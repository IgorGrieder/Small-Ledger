package handlers

import (
	"encoding/json"
	"fmt"
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
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("error decoding request json %v", request))
		return
	}

	RespondSuccess(w, http.StatusAccepted)
}
