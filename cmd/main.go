package main

import (
	"log/slog"
	"os"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting the program")

	// ENVs
	cfg := config.NewConfig()

	// Database connections
	connections := database.StartConns(cfg)

}
