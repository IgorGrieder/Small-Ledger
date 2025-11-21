package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
	"github.com/jackc/pgx/v5"
)

type LedgerService struct {
	store *repo.SQLStore
}

func NewLedgerService(store *repo.SQLStore) *LedgerService {
	return &LedgerService{
		store: store,
	}
}

var ErrNotEnoughFunds error = errors.New("not enough funds to proceed teh transaction")

func (l *LedgerService) InsertTransaction(ctx context.Context, transaction *domain.Transaction) error {
	// Using the same Tx for the whole processing
	dbTx := l.store.CreateTx()

	err := l.checkFunds(ctx, dbTx, transaction)
	if err != nil {
		slog.Error("error checking user funds",
			slog.String("error", err.Error()),
		)

		return err
	}

	return nil
}

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
