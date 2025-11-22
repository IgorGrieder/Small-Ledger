package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/IgorGrieder/Small-Ledger/internal/http/httpclient"
	"github.com/IgorGrieder/Small-Ledger/internal/http/httputils"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
	"github.com/jackc/pgx/v5"
)

type LedgerService struct {
	store      *repo.SQLStore
	httpClient *httpclient.Client
	cfg        *cfg.Config
}

type conversionRates struct {
	USD string `json:"USD"`
	BRL string `json:"BRL"`
}

type CurrencyResponse struct {
	ConversionRates conversionRates `json:"conversion_rates"`
}

var ErrNotEnoughFunds error = errors.New("not enough funds to proceed teh transaction")

func NewLedgerService(cfg *cfg.Config, store *repo.SQLStore, httpClient *httpclient.Client) *LedgerService {
	return &LedgerService{
		store:      store,
		httpClient: httpClient,
		cfg:        cfg,
	}
}

func (l *LedgerService) ProcessTransaction(ctx context.Context, transaction *domain.Transaction) error {
	// Using the same Tx for the whole processing
	dbTx := l.store.CreateTx()

	err := l.checkFunds(ctx, dbTx, transaction)
	if err != nil {
		slog.Error("error checking user funds",
			slog.String("error", err.Error()),
		)

		return err
	}

	err = l.checkCurrency(ctx, transaction)

	return nil
}

// Usuario pode quere transferir BRL para sambley
func (l *LedgerService) checkFunds(ctx context.Context, dbTx pgx.Tx, transaction *domain.Transaction) error {
	ctxQuery, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	tx := l.store.WithTx(dbTx)

	funds, err := tx.GetUserFunds(ctxQuery, transaction.From)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found %v", err)
		}

		return fmt.Errorf("error consulting user to check funds %v", err)
	}

	if funds < transaction.Value {
		return ErrNotEnoughFunds
	}

	return nil
}

func (l *LedgerService) checkCurrency(ctx context.Context, transaction *domain.Transaction) error {
	requests := []httpclient.ConcurrentRequest{
		{
			URL:    l.cfg.CURRENCY_URL + transaction.Currency,
			Method: http.MethodGet,
		},
		{
			URL:    l.cfg.CURRENCY_URL + transaction.Currency,
			Method: http.MethodGet,
		},
	}

	for result := range l.httpClient.FetchConcurrent(ctx, requests) {
		if result.Error != nil {
			return fmt.Errorf("request failed: error while checking currency %v", result.Error)
		}

		if result.Response.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed with status %d", result.Response.StatusCode)
		}
		defer result.Response.Body.Close()

		var response CurrencyResponse
		if err := httputils.DecodeJSONRaw(result.Response.Body, response); err != nil {
			return fmt.Errorf("request failed: desserializing json")
		}

	}

	return nil
}
