package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (api *APIServer) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	switchID := r.URL.Query().Get("switch_id")
	if switchID == "" {
		http.Error(w, "Missing switch_id parameter", http.StatusBadRequest)
		return
	}

	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Missing metric parameter", http.StatusBadRequest)
		return
	}

	val, err := api.dao.GetMetric(ctx, switchID, metricName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving metric: %v", err),
			http.StatusBadRequest)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the value to JSON (handles string, float, int, bool, etc.)
	jsonData, err := json.Marshal(val)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding value to JSON: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
