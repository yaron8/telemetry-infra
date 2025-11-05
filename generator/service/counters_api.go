package service

import (
	"fmt"
	"net/http"
)

// countersHandler handles the /counters endpoint
func (api *APIServer) countersHandler(w http.ResponseWriter, r *http.Request) {
	csvMetricsResponse, err := api.csvMetrics.GetCSVMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating CSV metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(csvMetricsResponse.HTTPResponseCode)
	fmt.Fprint(w, csvMetricsResponse.CSVData)
}
