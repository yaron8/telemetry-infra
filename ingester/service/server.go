package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yaron8/telemetry-infra/ingester/config"
	"github.com/yaron8/telemetry-infra/ingester/dao"
	"github.com/yaron8/telemetry-infra/telemetrics"
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
		w.Write([]byte("OK"))
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

func (api *APIServer) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	// Try to update metrics before serving
	if err := api.updateMetrics(); err != nil {
		http.Error(w, fmt.Sprintf("Error updating metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	allKeysAndMetrics, err := api.dao.GetAll(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the metrics to JSON
	jsonData, err := json.Marshal(allKeysAndMetrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding metrics to JSON: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (api *APIServer) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Try to update metrics before serving
	if err := api.updateMetrics(); err != nil {
		http.Error(w, fmt.Sprintf("Error updating metrics: %v", err),
			http.StatusInternalServerError)
		return
	}

	switchID := r.URL.Query().Get("switch_id")
	if switchID == "" {
		http.Error(w, "Missing switch_id parameter", http.StatusBadRequest)
		return
	}

	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Missing metric parameter", http.StatusBadRequest)
		return
	}

	val, err := api.dao.GetMetric(ctx, switchID, metricName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving metric: %v", err),
			http.StatusBadRequest)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the value to JSON (handles string, float, int, bool, etc.)
	jsonData, err := json.Marshal(val)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding value to JSON: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (api *APIServer) updateMetrics() error {
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
		api.WriteMetricsLineByLine(resp.Body)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (api *APIServer) WriteMetricsLineByLine(respBody io.ReadCloser) error {
	fmt.Println("Writing metrics..")

	scanner := bufio.NewScanner(respBody)
	ctx := context.Background()

	// Skip the header line
	if scanner.Scan() {
		// header line is skipped
	}

	lineNumber := 1
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Ignore empty lines (including lines with only whitespace)
		if line == "" {
			continue
		}

		// Parse the CSV line into a MetricRecord
		switchID, record, err := parseCSVLine(line)
		if err != nil {
			fmt.Printf("Error parsing line %d: %v\n", lineNumber, err)
			continue
		}

		// Store the metric using the DAO
		if err := api.dao.AddKey(ctx, switchID, record); err != nil {
			fmt.Printf("Error storing metric at line %d: %v\n", lineNumber, err)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	return nil
}

// parseCSVLine parses a CSV line into a MetricRecord
// Expected format: switch_id,bandwidth_mbps,latency_ms,packet_errors
func parseCSVLine(line string) (string, telemetrics.MetricRecord, error) {
	fields := strings.Split(line, ",")

	if len(fields) != 4 {
		return "", telemetrics.MetricRecord{}, fmt.Errorf("expected 4 fields, got %d", len(fields))
	}

	// Parse bandwidth_mbps
	bandwidthMbps, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
	if err != nil {
		return "", telemetrics.MetricRecord{}, fmt.Errorf("invalid bandwidth_mbps: %w", err)
	}

	// Parse latency_ms
	latencyMs, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
	if err != nil {
		return "", telemetrics.MetricRecord{}, fmt.Errorf("invalid latency_ms: %w", err)
	}

	// Parse packet_errors
	packetErrors, err := strconv.Atoi(strings.TrimSpace(fields[3]))
	if err != nil {
		return "", telemetrics.MetricRecord{}, fmt.Errorf("invalid packet_errors: %w", err)
	}

	switchID := strings.TrimSpace(fields[0])

	return switchID,
		telemetrics.MetricRecord{

			BandwidthMbps: bandwidthMbps,
			LatencyMs:     latencyMs,
			PacketErrors:  packetErrors,
		}, nil
}
