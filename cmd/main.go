package main

import (
	"log/slog"
	"os"

	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting the program")

	// ENVs
	cfg := cfg.NewConfig()

	// Database connections

}
