package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/yaron8/telemetry-infra/ingester/config"
	"github.com/yaron8/telemetry-infra/ingester/dao"
)

type APIServer struct {
	config *config.Config
	server *http.Server
	dao    *dao.DAOMetrics
}

func NewAPIServer(config *config.Config, dao *dao.DAOMetrics) *APIServer {

	return &APIServer{
		config: config,
		dao:    dao,
	}
}

// Start initializes and starts the HTTP server
func (api *APIServer) Start() error {

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			fmt.Printf("Error writing health check response: %v\n", err)
		}
	})

	// Telemetry endpoints
	mux.HandleFunc("/telemetry/ListMetrics", api.ListMetricsHandler)
	mux.HandleFunc("/telemetry/GetMetric", api.GetMetricHandler)

	api.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", api.config.Port),
		Handler:      mux,
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
