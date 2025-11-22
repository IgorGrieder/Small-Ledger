package httputils

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const InternalSrvErrMsg string = "error processing the request, try again"

type ErrorResponse struct {
	Error string `json:"error"`
}

func EncodeJsonErrorResponse(w http.ResponseWriter, status int, error ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(error); err != nil {
		slog.Error("failed to encode json response", "error", err)
	}
}

func RespondError(w http.ResponseWriter, status int, message string) {
	slog.Error("error while processing the request",
		slog.String("message", message),
		slog.Int("http code", status),
	)

	EncodeJsonErrorResponse(w, status, ErrorResponse{Error: message})
}

func RespondSuccess(w http.ResponseWriter, status int) {
	slog.Error("succeeded processing the request",
		slog.Int("http code", status),
	)

	w.WriteHeader(status)
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, request any) error {
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("error decoding request json %v", request))
		return err
	}

	return nil
}

func DecodeJSONRaw(source io.Reader, target any) error {
	if err := json.NewDecoder(source).Decode(&target); err != nil {
		return err
	}

	return nil
}
