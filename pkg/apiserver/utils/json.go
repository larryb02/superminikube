package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("content-type:", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "failed to process request", http.StatusInternalServerError)
		slog.Error("failed to write json response", "error", err)
	}
}