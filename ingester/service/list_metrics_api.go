package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (api *APIServer) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	allKeysAndMetrics, err := api.dao.GetAll(ctx)
	if err != nil {
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
		fmt.Printf("Error encoding metrics to JSON: %v\n", err)
		return
	}
}
