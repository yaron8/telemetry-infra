package service

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/yaron8/telemetry-infra/ingester/config"
	"github.com/yaron8/telemetry-infra/ingester/dao"
	"github.com/yaron8/telemetry-infra/logi"
)

type APIServer struct {
	config *config.Config
	server *http.Server
	dao    *dao.DAOMetrics
	logger *slog.Logger
}

func NewAPIServer(config *config.Config, dao *dao.DAOMetrics) *APIServer {

	return &APIServer{
		config: config,
		dao:    dao,
		logger: logi.GetLogger(),
	}
}

// Start initializes and starts the HTTP server
func (api *APIServer) Start() error {
	api.logger.Info("Ingester APIServer starting", "port", api.config.Port)

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			api.logger.Error("Error writing health check response", "error", err)
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

	if err := api.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		api.logger.Error("Server failed to start", "error", err, "port", api.config.Port)
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
