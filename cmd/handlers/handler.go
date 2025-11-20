package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/IgorGrieder/Small-Ledger/internal/application"
)

type LedgerHandler struct {
	ledgerService *application.LedgerService
}

func NewLedgerHandler(l *application.LedgerService) *LedgerHandler {
	return &LedgerHandler{ledgerService: l}
}

type transactionRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

func (h *LedgerHandler) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	var request transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("error decoding request json %v", request))
		return
	}

	err := h.ledgerService.InsertTransaction()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "error processing the request, try again")
	}
	RespondSuccess(w, http.StatusAccepted)
}
