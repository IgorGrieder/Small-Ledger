package handlers

import (
	"errors"
	"net/http"

	"github.com/IgorGrieder/Small-Ledger/internal/application"
	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/IgorGrieder/Small-Ledger/internal/http/httputils"
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
	Value    int64     `json:"value"`
}

func (h *LedgerHandler) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	var request transactionRequest
	err := httputils.DecodeJSON(w, r, &request)
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

	err = h.ledgerService.ProcessTransaction(r.Context(), transaction)
	if err != nil {

		if errors.Is(err, application.ErrNotEnoughFunds) {
			httputils.RespondError(w, http.StatusBadRequest, application.ErrNotEnoughFunds.Error())
			return
		}

		httputils.RespondError(w, http.StatusInternalServerError, httputils.InternalSrvErrMsg)
	}

	httputils.RespondSuccess(w, http.StatusAccepted)
}
