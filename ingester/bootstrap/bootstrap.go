package bootstrap

import (
	"fmt"
	"net/http"
	"time"

	"github.com/yaron8/telemetry-infra/ingester/config"
)

type Bootstrap struct {
	config *config.Config
	server *http.Server
}

func NewBootstrap() (*Bootstrap, error) {
	// Load configuration
	cfg := config.NewConfig()

	return &Bootstrap{
		config: cfg,
	}, nil
}

// Start initializes and starts the HTTP server
func (b *Bootstrap) Start() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Telemetry endpoints
	mux.HandleFunc("/telemetry/ListMetrics", ListMetricsHandler)
	mux.HandleFunc("/telemetry/GetMetric", GetMetricHandler)

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

func ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello ListMetrics"))
}

func GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	switchID := r.URL.Query().Get("switch_id")
	metric := r.URL.Query().Get("metric")

	response := fmt.Sprintf("switch_id: %s, metric: %s", switchID, metric)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
