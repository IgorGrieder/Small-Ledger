package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerService struct {
	repository *repo.Queries
	conn       *pgxpool.Pool
}

func NewLedgerService(conn *pgxpool.Pool) *LedgerService {
	repo := repo.New(conn)

	return &LedgerService{repository: repo,
		conn: conn,
	}
}

var ErrNotEnoughFunds error = errors.New("not enough funds to proceed teh transaction")

func (l *LedgerService) InsertTransaction(ctx context.Context, transaction *domain.Transaction) error {
	err := l.checkFunds(ctx, transaction)
	if err != nil {
		slog.Error("error checking user funds",
			slog.String("error", err.Error()),
		)

		return err
	}

	return nil
}

func (l *LedgerService) checkFunds(ctx context.Context, transaction *domain.Transaction) error {
	ctxQuery, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	tx, err := l.conn.Begin(ctxQuery)
	if err != nil {
		return fmt.Errorf("error opening a transaction %v", err)
	}

	querier := l.repository.WithTx(tx)

	funds, err := querier.GetUserFunds(ctxQuery, transaction.From)
	if err != nil {
		return fmt.Errorf("error checking user funds %v", err)
	}

	if funds < transaction.Value {
		return ErrNotEnoughFunds
	}

	return err
}
