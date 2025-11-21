package repo

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

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

func (s *SQLStore) CreateTx() *pgxpool.Tx {
	return &pgxpool.Tx{}
}
