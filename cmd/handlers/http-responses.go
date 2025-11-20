package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func EncodeJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode json response", "error", err)
	}
}

func RespondError(w http.ResponseWriter, status int, message string) {
	slog.Error("error while processing the request",
		slog.String("message", message),
		slog.Int("http code", status),
	)

	EncodeJSONResponse(w, status, ErrorResponse{Error: message})
}

func RespondSuccess(w http.ResponseWriter, status int) {
	slog.Error("succeeded processing the request",
		slog.Int("http code", status),
	)

	w.WriteHeader(status)
}

