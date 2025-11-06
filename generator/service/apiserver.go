package service

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/yaron8/telemetry-infra/generator/config"
	"github.com/yaron8/telemetry-infra/generator/metrics"
	"github.com/yaron8/telemetry-infra/logi"
)

type APIServer struct {
	csvMetrics *metrics.CSVMetrics
	config     *config.Config
	server     *http.Server
	logger     *slog.Logger
}

func NewAPIServer(config *config.Config, csvMetrics *metrics.CSVMetrics) *APIServer {
	return &APIServer{
		config:     config,
		csvMetrics: csvMetrics,
		logger:     logi.GetLogger(),
	}
}

// StartServer initializes and starts the HTTP server
func (api *APIServer) Start() error {
	api.logger.Info("APIServer starting", "port", api.config.Port)

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			fmt.Printf("Error writing health check response: %v\n", err)
		}
	})

	// Set up HTTP handlers
	mux.HandleFunc("/counters", api.countersHandler)

	// Wrap the mux with logging middleware
	handler := api.middleware(mux)

	api.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", api.config.Port),
		Handler:      handler,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Starting server on port %d\n", api.config.Port)
	if err := api.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
