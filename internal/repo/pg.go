package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func SetupPg() *Queries {

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "user=pqgotest dbname=pqgotest sslmode=verify-full")
	if err != nil {
		return nil
	}
	defer conn.Close(ctx)

	return New(conn)
}
