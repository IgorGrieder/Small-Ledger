package application

import (
	"github.com/IgorGrieder/Small-Ledger/internal/domain"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
)

type LedgerService struct {
	repository repo.Querier
}

func NewLedgerService(r repo.Querier) *LedgerService {
	return &LedgerService{repository: r}
}

func (l *LedgerService) InsertTransaction(transaction *domain.Transaction) error {
	return nil
}
