package repo

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateTx(ctx context.Context) (pgx.Tx, error)
	QueryConcurrent(ctx context.Context, queries []ConcurrentQuery) <-chan QueryResponse
}

type ConcurrentQuery struct {
	Fn func(context.Context) (any, error)
}

type QueryResponse struct {
	Name   string
	Result any
	Error  error
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

func (s *SQLStore) QueryConcurrent(ctx context.Context, queries []ConcurrentQuery) <-chan QueryResponse {
	results := make(chan QueryResponse, len(queries))
	var wg sync.WaitGroup

	for _, q := range queries {
		wg.Add(1)
		go func(query ConcurrentQuery) {
			defer wg.Done()
			res, err := query.Fn(ctx)
			results <- QueryResponse{
				Result: res,
				Error:  err,
			}
		}(q)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
