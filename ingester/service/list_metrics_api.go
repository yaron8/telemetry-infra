package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (api *APIServer) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	api.logger.Info("ListMetricsHandler called")

	allKeysAndMetrics, err := api.dao.GetAll(ctx)
	if err != nil {
		api.logger.Error("Error retrieving metrics", "error", err)
		http.Error(w, fmt.Sprintf("Error retrieving metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	// Set content type and status code before encoding
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Use encoder to stream JSON directly to response writer
	// This avoids allocating the entire JSON in memory
	if err := json.NewEncoder(w).Encode(allKeysAndMetrics); err != nil {
		// Can't send error response after WriteHeader, just log it
		api.logger.Error("Error encoding metrics to JSON", "error", err)
		return
	}
}
