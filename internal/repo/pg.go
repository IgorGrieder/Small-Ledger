package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupPg() *pgxpool.Pool {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx, "user=pqgotest dbname=pqgotest sslmode=verify-full")
	if err != nil {
		return nil
	}

	return conn
}
