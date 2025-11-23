package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/IgorGrieder/Small-Ledger/internal/application"
	"github.com/IgorGrieder/Small-Ledger/internal/cfg"
)

func StartServer(ledger *application.LedgerService, cfg *cfg.Config) {
	ledgerHandler := NewLedgerHandler(ledger)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /transaction", ledgerHandler.TransactionHandler)
	mux.HandleFunc("GET /accounts", ledgerHandler.GetAccountsHandler)

	srv := &http.Server{Addr: fmt.Sprintf("localhost:%d", cfg.APPLICATION_PORT), Handler: mux}

	if err := srv.ListenAndServe(); err != nil {
		log.Println("Server stopped")
		os.Exit(1)
	}
}
