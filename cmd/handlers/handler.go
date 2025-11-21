package handlers

import (
	"errors"
	"net/http"

	"github.com/IgorGrieder/Small-Ledger/internal/application"
	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/google/uuid"
)

type LedgerHandler struct {
	ledgerService *application.LedgerService
}

func NewLedgerHandler(l *application.LedgerService) *LedgerHandler {
	return &LedgerHandler{ledgerService: l}
}

type transactionRequest struct {
	From     uuid.UUID `json:"from"`
	To       uuid.UUID `json:"to"`
	Currency string    `json:"currency"`
	Value    string    `json:"value"`
}

func (h *LedgerHandler) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	var request transactionRequest
	err := DecodeJSON(w, r, &request)
	if err != nil {
		return
	}

	transaction := &domain.Transaction{
		From:          request.From,
		To:            request.To,
		Currency:      request.Currency,
		Value:         request.Value,
		CorrelationId: uuid.New(),
	}

	err = h.ledgerService.InsertTransaction(r.Context(), transaction)
	if err != nil {

		if errors.Is(err, application.ErrNotEnoughFunds) {
			RespondError(w, http.StatusBadRequest, "not enouth funds to process the transaction")
			return
		}

		RespondError(w, http.StatusInternalServerError, "error processing the request, try again")
	}

	RespondSuccess(w, http.StatusAccepted)
}
