package application

import (
	"context"
	"errors"
	"time"

	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
)

type LedgerService struct {
	repository repo.Querier
}

func NewLedgerService(r repo.Querier) *LedgerService {
	return &LedgerService{repository: r}
}

var ErrNotEnoughFunds error = errors.New("not enough funds to proceed teh transaction")

func (l *LedgerService) InsertTransaction(ctx context.Context, transaction *domain.Transaction) error {
	funds, err := checkFunds(ctx, transaction)

	return nil
}

func (l *LedgerService) checkFunds(ctx context.Context, transaction *domain.Transaction) (int64, error) {
	ctxQuery, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	funds, err := l.repository.GetUserFunds(ctxQuery, transaction.From)
	if err != nil {

	}
	return funds, err
}
