package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/IgorGrieder/Small-Ledger/cmd/handlers"
	"github.com/IgorGrieder/Small-Ledger/internal/application"
	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
	"github.com/IgorGrieder/Small-Ledger/internal/http/httpclient"
	"github.com/IgorGrieder/Small-Ledger/internal/repo"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting the program")

	// ENVs
	cfg := cfg.NewConfig()

	// Redis and Pg
	pgConn := repo.SetupPg(cfg)
	redis := repo.SetupRedis(cfg)
	store := repo.NewStore(pgConn)

	defer pgConn.Close()
	defer redis.Close()

	// Http base client
	httpClient := httpclient.NewClient(60*time.Second, 5, 10*time.Second)

	ledgerService := application.NewLedgerService(cfg, store, redis, httpClient)

	handlers.StartServer(ledgerService, cfg)
}
