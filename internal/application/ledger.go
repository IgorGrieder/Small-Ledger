package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/IgorGrieder/Small-Ledger/internal/http/httpclient"
	"github.com/IgorGrieder/Small-Ledger/internal/http/httputils"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type LedgerService struct {
	store      *repo.SQLStore
	httpClient *httpclient.Client
	cfg        *cfg.Config
	redis      *redis.Client
}

type conversionRates struct {
	USD string `json:"USD"`
	BRL string `json:"BRL"`
}

type CurrencyResponse struct {
	ConversionRates conversionRates `json:"conversion_rates"`
}

var ErrNotEnoughFunds error = errors.New("not enough funds to proceed teh transaction")

func NewLedgerService(cfg *cfg.Config, store *repo.SQLStore, redis *redis.Client, httpClient *httpclient.Client) *LedgerService {
	return &LedgerService{
		store:      store,
		httpClient: httpClient,
		cfg:        cfg,
		redis:      redis,
	}
}

func (l *LedgerService) ProcessTransaction(ctx context.Context, transaction *domain.Transaction) error {
	currency, err := l.checkCurrency(ctx, transaction)
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

func (l *LedgerService) checkFunds(ctx context.Context, queries *repo.Queries, transaction *domain.Transaction) error {
	ctxQuery, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var (
		fundsFrom int64
		fundsTo   int64
	)

	group, ctxGroup := errgroup.WithContext(ctxQuery)

	fetchFunds := func(userID uuid.UUID, label string, dest *int64) error {
		funds, err := l.store.GetUserFunds(ctxGroup, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("user not found for %s=%s: %w", label, userID, err)
			}
			return fmt.Errorf("error consulting user %s=%s: %w", label, userID, err)
		}
		*dest = funds
		return nil
	}

	group.Go(func() error {
		return fetchFunds(transaction.From, "From", &fundsFrom)
	})

	group.Go(func() error {
		return fetchFunds(transaction.To, "To", &fundsTo)
	})

	if err := group.Wait(); err != nil {
		return err
	}

	if fundsFrom < transaction.Value {
		return ErrNotEnoughFunds
	}

	return nil
}

func (l *LedgerService) checkCurrency(ctx context.Context, transaction *domain.Transaction) (conversionRates, error) {
	response, err := l.httpClient.Get(ctx, l.cfg.CURRENCY_URL+transaction.Currency, nil, nil)
	if err != nil {
		return conversionRates{}, fmt.Errorf("error checking currency %v", err)
	}

	var rates CurrencyResponse
	httputils.DecodeJSONRaw(response.Body, rates)

	return conversionRates{USD: rates.ConversionRates.USD, BRL: rates.ConversionRates.BRL}, nil
}
