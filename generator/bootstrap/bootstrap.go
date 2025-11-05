package bootstrap

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yaron8/telemetry-infra/generator/metrics"
)

const serverPort = ":9001"

type Bootstrap struct {
	csvMetrics *metrics.CSVMetrics
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{
		csvMetrics: metrics.NewCSVMetrics(),
	}
}

// StartServer initializes and starts the HTTP server on port 9001
func (b *Bootstrap) StartServer() error {
	// Set up HTTP handlers
	http.HandleFunc("/counters", b.countersHandler)

	// Start the server
	fmt.Printf("Starting HTTP server on port %s\n", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return err
	}

	return nil
}

// countersHandler handles the /counters endpoint
func (b *Bootstrap) countersHandler(w http.ResponseWriter, r *http.Request) {
	csvData, err := b.csvMetrics.GetCSVMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating CSV metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	fmt.Fprint(w, csvData)
}
