package metrics

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/yaron8/telemetry-infra/metrics"
)

const numOfDataLines = 100

type CSVMetrics struct {
	mu                      sync.RWMutex
	snapshot                string
	snapshotLastTimeUpdated time.Time
	snapshotTTL             time.Duration
}

func NewCSVMetrics(snapshotTTL time.Duration) *CSVMetrics {
	return &CSVMetrics{
		snapshotTTL: snapshotTTL,
	}
}

// GetCSVMetrics generates CSV formatted metrics data with caching
func (cm *CSVMetrics) GetCSVMetrics() (string, error) {
	// Check if snapshot is valid
	cm.mu.RLock()
	if time.Since(cm.snapshotLastTimeUpdated) < cm.snapshotTTL && cm.snapshot != "" {
		snapshot := cm.snapshot
		cm.mu.RUnlock()
		return snapshot, nil
	}
	cm.mu.RUnlock()

	// Snapshot is expired or empty, generate new data
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have updated it)
	if time.Since(cm.snapshotLastTimeUpdated) < cm.snapshotTTL && cm.snapshot != "" {
		return cm.snapshot, nil
	}

	// Create a buffer to write CSV data to
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"timestamp", "switch_ame", "bandwidth_mbps", "latency_ms", "packet_errors"}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("error writing header: %w", err)
	}

	// Get current timestamp
	currentTime := time.Now().Unix()

	// Generate numOfDataLines lines of data
	for i := 1; i <= numOfDataLines; i++ {
		// Generate random metrics data
		metric := metrics.Metric{
			Timestamp:     currentTime,
			SwitchName:    fmt.Sprintf("sw%d", i),
			BandwidthMbps: rand.Float64() * 10000, // Random bandwidth up to 10 Gbps
			LatencyMs:     rand.Float64() * 100,   // Random latency up to 100ms
			PacketErrors:  rand.Intn(1000),        // Random packet errors up to 1000
		}

		row := []string{
			fmt.Sprintf("%d", metric.Timestamp),
			metric.SwitchName,
			fmt.Sprintf("%.2f", metric.BandwidthMbps),
			fmt.Sprintf("%.2f", metric.LatencyMs),
			fmt.Sprintf("%d", metric.PacketErrors),
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("error writing row: %w", err)
		}
	}

	// Flush the writer to ensure all data is written to the buffer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("error flushing writer: %w", err)
	}

	// Save to snapshot
	cm.snapshot = buf.String()
	cm.snapshotLastTimeUpdated = time.Now()

	return cm.snapshot, nil
}
