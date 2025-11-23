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
	_, err := l.checkCurrency(ctx, transaction)
	if err != nil {
		slog.Error("error checking currency",
			slog.String("error", err.Error()),
		)

		return err
	}

	tx, err := l.store.CreateTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := l.store.WithTx(tx)

	err = l.checkFunds(ctx, qtx, transaction)
	if err != nil {
		slog.Error("error checking user funds",
			slog.String("error", err.Error()),
		)

		return err
	}

	return tx.Commit(ctx)
}

// Usuario pode quere transferir BRL para sambley
func (l *LedgerService) checkFunds(ctx context.Context, queries *repo.Queries, transaction *domain.Transaction) error {
	ctxQuery, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	queriesParallel := []repo.ConcurrentQuery{
		{
			Name: "From",
			Fn: func(ctx context.Context) (any, error) {
				return l.store.GetUserFunds(ctxQuery, transaction.From)
			},
		},
		{
			Name: "To",
			Fn: func(ctx context.Context) (any, error) {
				return l.store.GetUserFunds(ctxQuery, transaction.To)
			},
		},
	}

	var fundsFrom int64
	// var fundsTo int64

	for res := range l.store.QueryConcurrent(ctx, queriesParallel) {
		if res.Error != nil {
			if errors.Is(res.Error, pgx.ErrNoRows) {
				return fmt.Errorf("user not found for %s: %w", res.Name, res.Error)
			}
			return fmt.Errorf("error consulting user %s to check funds: %w", res.Name, res.Error)
		}

		val, ok := res.Result.(int64)
		if !ok {
			return fmt.Errorf("unexpected result type for %s", res.Name)
		}

		if res.Name == "From" {
			fundsFrom = val
		} else if res.Name == "To" {
			// fundsTo = val
		}
	}

	if fundsFrom < transaction.Value {
		return ErrNotEnoughFunds
	}

	// fundsTo is fetched but not strictly used for a check here, keeping it as per original logic intent (maybe for future use or just existence check)

	return nil
}

func (l *LedgerService) checkCurrency(ctx context.Context, transaction *domain.Transaction) ([]conversionRates, error) {
	var response CurrencyResponse
	var rates []conversionRates
	requests := []httpclient.ConcurrentRequest{
		{
			URL:    l.cfg.CURRENCY_URL + transaction.Currency,
			Method: http.MethodGet,
		},
	}

	for result := range l.httpClient.FetchConcurrent(ctx, requests) {
		if result.Error != nil {
			return nil, fmt.Errorf("request failed: error while checking currency %v", result.Error)
		}

		if result.Response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed with status %d", result.Response.StatusCode)
		}
		defer result.Response.Body.Close()

		if err := httputils.DecodeJSONRaw(result.Response.Body, response); err != nil {
			return nil, fmt.Errorf("request failed: desserializing json")
		}

		rates = append(rates, response.ConversionRates)
	}

	return rates, nil
}
