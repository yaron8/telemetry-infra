package service

import (
	"fmt"
	"net/http"
)

// countersHandler handles the /counters endpoint
func (api *APIServer) countersHandler(w http.ResponseWriter, r *http.Request) {
	api.logger.Info("countersHandler called")

	csvMetricsResponse, err := api.csvMetrics.GetCSVMetrics()
	if err != nil {
		api.logger.Error("Error generating CSV metrics", "error", err)
		http.Error(w, fmt.Sprintf("Error generating CSV metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(csvMetricsResponse.HTTPResponseCode)
	fmt.Fprint(w, csvMetricsResponse.CSVData)
}
