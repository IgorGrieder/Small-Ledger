package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/IgorGrieder/Small-Ledger/cmd/handlers"
	"github.com/IgorGrieder/Small-Ledger/internal/application"
	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
)

func StartServer(ledger *application.LedgerService, cfg *cfg.Config) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /transaction", handlers.TransactionHandler)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.PORT), Handler: mux}

	if err := srv.ListenAndServe(); err != nil {
		log.Println("Server stopped")
		os.Exit(1)
	}
}
