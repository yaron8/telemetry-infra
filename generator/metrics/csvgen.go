package metrics

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/yaron8/telemetry-infra/logi"
	"github.com/yaron8/telemetry-infra/telemetrics"
)

const numOfDataLines = 100

type CSVMetrics struct {
	mu                      sync.RWMutex
	snapshotLastTimeUpdated time.Time
	snapshotTTL             time.Duration
	logger                  *slog.Logger
}

type CSVMetricsResponse struct {
	CSVData          string
	HTTPResponseCode int
}

func NewCSVMetrics(snapshotTTL time.Duration) *CSVMetrics {
	return &CSVMetrics{
		snapshotTTL: snapshotTTL,
		logger:      logi.GetLogger(),
	}
}

// GetCSVMetrics generates CSV formatted metrics data with caching
func (cm *CSVMetrics) GetCSVMetrics() (*CSVMetricsResponse, error) {
	// Check if snapshot is valid (hot path - no logging for performance)
	cm.mu.RLock()
	if time.Since(cm.snapshotLastTimeUpdated) < cm.snapshotTTL {
		cm.mu.RUnlock()
		return &CSVMetricsResponse{
			CSVData:          "",
			HTTPResponseCode: http.StatusNotModified,
		}, nil
	}
	cm.mu.RUnlock()

	// Snapshot is expired or empty, generate new data
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have updated it)
	if time.Since(cm.snapshotLastTimeUpdated) < cm.snapshotTTL {
		return &CSVMetricsResponse{
			CSVData:          "",
			HTTPResponseCode: http.StatusOK,
		}, nil
	}

	// Only log when actually generating new data (cold path)
	cm.logger.Info("Generating new CSV metrics", "num_lines", numOfDataLines)

	// Create a buffer to write CSV data to
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := telemetrics.GetCSVHeader()
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("error writing header: %w", err)
	}

	currTimestamp := time.Now().Unix()

	// Generate numOfDataLines lines of data
	for i := 1; i <= numOfDataLines; i++ {
		// Generate random metrics data
		metric := telemetrics.MetricRecord{
			Timestamp:     currTimestamp,
			SwitchID:      fmt.Sprintf("sw%d", i),
			BandwidthMbps: rand.Float64() * 10000, // Random bandwidth up to 10 Gbps
			LatencyMs:     rand.Float64() * 5000,  // Random latency up to 5 seconds
			PacketErrors:  rand.Intn(100),         // Random packet errors up to 1000
		}

		row := []string{
			fmt.Sprintf("%d", metric.Timestamp),
			metric.SwitchID,
			fmt.Sprintf("%.2f", metric.BandwidthMbps),
			fmt.Sprintf("%.2f", metric.LatencyMs),
			fmt.Sprintf("%d", metric.PacketErrors),
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("error writing row: %w", err)
		}
	}

	// Flush the writer to ensure all data is written to the buffer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("error flushing writer: %w", err)
	}

	// Save to snapshot
	snapshot := buf.String()
	cm.snapshotLastTimeUpdated = time.Now()

	cm.logger.Info("CSV metrics generated successfully",
		"data_size_bytes", len(snapshot),
		"num_lines", numOfDataLines,
		"timestamp", currTimestamp)

	return &CSVMetricsResponse{
		CSVData:          snapshot,
		HTTPResponseCode: http.StatusOK,
	}, nil
}
