package main

import (
	"log/slog"
	"os"

	"github.com/IgorGrieder/Small-Ledger/internal/application"
	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting the program")

	// ENVs
	cfg := cfg.NewConfig()

	// Redis and Pg
	pgConn := repo.SetupPg()
	defer pgConn.Close()

	ledgerService := application.NewLedgerService(pgConn)

	StartServer(ledgerService, cfg)
}
