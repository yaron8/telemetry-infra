package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (api *APIServer) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	api.logger.Info("GetMetricHandler called")

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
		api.logger.Error("Error getting metric", "switch_id", switchID, "metric", metricName, "error", err)
		http.Error(w, err.Error(),
			http.StatusNotFound)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the value to JSON (handles string, float, int, bool, etc.)
	jsonData, err := json.Marshal(val)
	if err != nil {
		api.logger.Error("Error encoding value to JSON", "switch_id", switchID, "metric", metricName, "error", err)
		http.Error(w, fmt.Sprintf("Error encoding value to JSON: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonData); err != nil {
		api.logger.Error("Error writing response", "error", err)
	}
}
