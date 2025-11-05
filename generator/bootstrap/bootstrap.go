package bootstrap

import (
	"fmt"
	"net/http"
	"time"

	"github.com/yaron8/telemetry-infra/generator/config"
	"github.com/yaron8/telemetry-infra/generator/metrics"
)

type Bootstrap struct {
	csvMetrics *metrics.CSVMetrics
	config     *config.Config
	server     *http.Server
}

func NewBootstrap() (*Bootstrap, error) {
	// Load configuration
	cfg := config.NewConfig()

	return &Bootstrap{
		csvMetrics: metrics.NewCSVMetrics(cfg.SnapshotTTL),
		config:     cfg,
	}, nil
}

// StartServer initializes and starts the HTTP server
func (b *Bootstrap) StartServer() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Set up HTTP handlers
	mux.HandleFunc("/counters", b.countersHandler)

	b.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", b.config.Port),
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Starting server on port %d\n", b.config.Port)
	if err := b.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// countersHandler handles the /counters endpoint
func (b *Bootstrap) countersHandler(w http.ResponseWriter, r *http.Request) {
	csvMetricsResponse, err := b.csvMetrics.GetCSVMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating CSV metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(csvMetricsResponse.HTTPResponseCode)
	fmt.Fprint(w, csvMetricsResponse.CSVData)
}
