package bootstrap

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yaron8/telemetry-infra/ingester/config"
	"github.com/yaron8/telemetry-infra/ingester/metrics"
)

type Bootstrap struct {
	config  *config.Config
	server  *http.Server
	metrics metrics.Metrics
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
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", b.config.Redis.Host, b.config.Redis.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	b.metrics = *metrics.NewMetrics(redisClient, b.config.Redis.TTL)

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Telemetry endpoints
	mux.HandleFunc("/telemetry/ListMetrics", b.ListMetricsHandler)
	mux.HandleFunc("/telemetry/GetMetric", b.GetMetricHandler)

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

func (b *Bootstrap) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Try to update metrics before serving
	if err := b.updateMetrics(); err != nil {
		http.Error(w, fmt.Sprintf("Error updating metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello ListMetrics"))
}

func (b *Bootstrap) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	// Try to update metrics before serving
	if err := b.updateMetrics(); err != nil {
		http.Error(w, fmt.Sprintf("Error updating metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	switchID := r.URL.Query().Get("switch_id")
	metric := r.URL.Query().Get("metric")

	response := fmt.Sprintf("switch_id: %s, metric: %s", switchID, metric)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func (b *Bootstrap) updateMetrics() error {
	resp, err := http.Get("http://localhost:9001/counters")
	if err != nil {
		fmt.Println("error")
		return fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		fmt.Println("304 - skip")
	case http.StatusOK:
		b.WriteMetricsLineByLine(resp.Body)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (b *Bootstrap) WriteMetricsLineByLine(respBody io.ReadCloser) error {
	fmt.Println("Writing metrics..")
	return nil
}
