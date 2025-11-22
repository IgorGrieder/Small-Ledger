package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/IgorGrieder/Small-Ledger/internal/httpclient"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
	"github.com/jackc/pgx/v5"
)

type LedgerService struct {
	store      *repo.SQLStore
	httpClient *httpclient.Client
	cfg        *cfg.Config
}

func NewLedgerService(cfg *cfg.Config, store *repo.SQLStore, httpClient *httpclient.Client) *LedgerService {
	return &LedgerService{
		store:      store,
		httpClient: httpClient,
		cfg:        cfg,
	}
}

var ErrNotEnoughFunds error = errors.New("not enough funds to proceed teh transaction")

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
	ctxHttp, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	urls := []string{l.cfg.CURRENCY_URL, l.cfg.CURRENCY_URL}
	ch1 := make(chan httpclient.HTTPResponse)
	ch2 := make(chan httpclient.HTTPResponse)
	emptyMap := make(map[string]map[string]string)
	urlsMap := make(map[string]chan httpclient.HTTPResponse)
	urlsMap[urls[0]] = ch1
	urlsMap[urls[1]] = ch2

	var wg sync.WaitGroup

	l.httpClient.FetchConcurrentUrls(ctxHttp, &wg, urlsMap, http.MethodGet, emptyMap, emptyMap)

	wg.Wait()

	return nil
}
