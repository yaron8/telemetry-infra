package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (api *APIServer) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	// Try to update metrics before serving
	if err := api.updateMetrics(); err != nil {
		http.Error(w, fmt.Sprintf("Error updating metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	allKeysAndMetrics, err := api.dao.GetAll(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the metrics to JSON
	jsonData, err := json.Marshal(allKeysAndMetrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding metrics to JSON: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
