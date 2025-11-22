package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateTx(ctx context.Context) (pgx.Tx, error)
}

type SQLStore struct {
	*Queries
	connPool *pgxpool.Pool
}

func NewStore(connPool *pgxpool.Pool) *SQLStore {
	return &SQLStore{
		Queries:  New(connPool),
		connPool: connPool,
	}
}

func (s *SQLStore) CreateTx(ctx context.Context) (pgx.Tx, error) {
	return s.connPool.Begin(ctx)
}
