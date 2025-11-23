package application

import (
	"context"
	"encoding/json"
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
	cacheMax   time.Duration
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
		cacheMax:   1 * time.Hour,
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

func (l *LedgerService) GetAllAccounts(ctx context.Context) ([]repo.Account, error) {
	return l.store.GetAllAccounts(ctx)
}

func (l *LedgerService) checkFunds(ctx context.Context, tx *repo.Queries, transaction *domain.Transaction) error {
	ctxQuery, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var (
		fundsFrom int64
		fundsTo   int64
	)

	group, ctxGroup := errgroup.WithContext(ctxQuery)

	group.Go(func() error {
		return fetchFunds(tx, ctxGroup, transaction.From, "From", &fundsFrom)
	})

	group.Go(func() error {
		return fetchFunds(tx, ctxGroup, transaction.To, "To", &fundsTo)
	})

	if err := group.Wait(); err != nil {
		return err
	}

	if fundsFrom < transaction.Value {
		return ErrNotEnoughFunds
	}

	return nil
}

func fetchFunds(tx *repo.Queries, ctx context.Context, userID uuid.UUID, label string, dest *int64) error {
	funds, err := tx.GetUserFunds(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found for %s=%s: %w", label, userID, err)
		}
		return fmt.Errorf("error consulting user %s=%s: %w", label, userID, err)
	}
	*dest = funds
	return nil
}

func (l *LedgerService) checkCurrency(ctx context.Context, transaction *domain.Transaction) (conversionRates, error) {
	val, err := l.redis.Get(ctx, transaction.Currency).Result()
	if err == nil {
		var cachedRates conversionRates
		if err := json.Unmarshal([]byte(val), &cachedRates); err == nil {
			return cachedRates, nil
		}
	} else if err != redis.Nil {
		slog.Error("error getting currency from cache", slog.String("error", err.Error()))
	}

	response, err := l.httpClient.Get(ctx, l.cfg.CURRENCY_URL+transaction.Currency, nil, nil)
	if err != nil {
		return conversionRates{}, fmt.Errorf("error checking currency %v", err)
	}

	var rates CurrencyResponse
	if err := httputils.DecodeJSONRaw(response.Body, rates); err != nil {
		return conversionRates{}, fmt.Errorf("error decoding currency response: %w", err)
	}
	defer response.Body.Close()

	resultRates := conversionRates{USD: rates.ConversionRates.USD, BRL: rates.ConversionRates.BRL}
	if jsonData, err := json.Marshal(resultRates); err == nil {
		if err := l.redis.Set(ctx, transaction.Currency, jsonData, l.cacheMax).Err(); err != nil {
			slog.Error("error setting currency to cache", slog.String("error", err.Error()))
		}
	}

	return resultRates, nil
}
