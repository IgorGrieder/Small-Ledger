package repo

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupPg(cfg *cfg.Config) *pgxpool.Pool {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx,
		fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=prefer",
			cfg.PG_USER,
			cfg.PG_DB_NAME,
			cfg.PG_PASS,
			cfg.PG_HOST),
	)

	err = conn.Ping(ctx)
	if err != nil {
		slog.Error("failed connecting into Postgre",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	return conn
}
