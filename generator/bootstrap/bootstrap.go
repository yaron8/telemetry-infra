package bootstrap

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yaron8/telemetry-infra/generator/config"
	"github.com/yaron8/telemetry-infra/generator/metrics"
)

type Bootstrap struct {
	csvMetrics *metrics.CSVMetrics
	config     *config.Config
}

func NewBootstrap() (*Bootstrap, error) {
	// Load configuration
	cfg := config.NewConfig()

	return &Bootstrap{
		csvMetrics: metrics.NewCSVMetrics(cfg.CacheTTL),
		config:     cfg,
	}, nil
}

// StartServer initializes and starts the HTTP server on port 9001
func (b *Bootstrap) StartServer() error {
	// Set up HTTP handlers
	http.HandleFunc("/counters", b.countersHandler)

	serverPort := fmt.Sprintf(":%d", b.config.Port)
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
