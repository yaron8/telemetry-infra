package etl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yaron8/telemetry-infra/ingester/dao"
	"github.com/yaron8/telemetry-infra/logi"
	"github.com/yaron8/telemetry-infra/telemetrics"
)

type ETL struct {
	dao          *dao.DAOMetrics
	interval     time.Duration
	generatorURL string
	logger       *slog.Logger
}

func NewETL(dao *dao.DAOMetrics, interval time.Duration, generatorURL string) *ETL {
	return &ETL{
		dao:          dao,
		interval:     interval,
		generatorURL: generatorURL,
		logger:       logi.GetLogger(),
	}
}

func (etl *ETL) Run() {
	etl.logger.Info("ETL starting", "interval", etl.interval, "generator_url", etl.generatorURL)
	for {
		if err := etl.updateMetrics(); err != nil {
			etl.logger.Error("Error updating metrics", "error", err)
		}

		// Sleep until the next interval
		time.Sleep(etl.interval)
	}
}

func (etl *ETL) updateMetrics() error {
	resp, err := http.Get(etl.generatorURL + "/counters")
	if err != nil {
		etl.logger.Error("Error fetching metrics from generator in EP /counters", "error", err)
		return fmt.Errorf("failed to fetch metrics: %w", err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		// No logging on hot path - cache hit is normal
		return nil
	case http.StatusOK:
		etl.logger.Info("Fetching new metrics from generator")
		if err := etl.writeMetricsLineByLine(resp.Body); err != nil {
			return fmt.Errorf("failed to write metrics: %w", err)
		}
	default:
		etl.logger.Error("Unexpected status code from generator", "status_code", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (etl *ETL) writeMetricsLineByLine(respBody io.ReadCloser) error {
	scanner := bufio.NewScanner(respBody)
	ctx := context.Background()

	// Skip the header line
	if !scanner.Scan() {
		// No data to read
		return nil
	}

	lastTimeUpdated := int64(0)
	lineNumber := 1
	errorCount := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Ignore empty lines (including lines with only whitespace)
		if line == "" {
			continue
		}

		// Parse the CSV line into a MetricRecord
		switchID, timestamp, record, err := etl.parseCSVLine(line)
		if err != nil {
			errorCount++
			etl.logger.Error("Error parsing line", "line_number", lineNumber, "error", err)
			continue
		}

		// Store the metric using the DAO
		if err := etl.dao.AddMetric(ctx, timestamp, switchID, record); err != nil {
			errorCount++
			etl.logger.Error("Error storing metric", "line_number", lineNumber, "switch_id", switchID, "error", err)
			continue
		}

		if timestamp > 0 {
			lastTimeUpdated = timestamp
		}
	}

	// Update key in Redis for last update time
	if lastTimeUpdated > 0 {
		if err := etl.dao.SetLastUpdateTime(ctx, lastTimeUpdated); err != nil {
			return fmt.Errorf("failed to set last update time: %w", err)
		}
		etl.logger.Info("Metrics processed successfully",
			"total_lines", lineNumber-1,
			"errors", errorCount,
			"last_timestamp", lastTimeUpdated)
	} else {
		etl.logger.Error("No valid timestamp found to update last update time")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	return nil
}

// parseCSVLine parses a CSV line into a MetricRecord
// Expected format: switch_id,bandwidth_mbps,latency_ms,packet_errors
func (etl *ETL) parseCSVLine(line string) (string, int64, telemetrics.MetricRecord, error) {
	fields := strings.Split(line, ",")

	if len(fields) != 5 {
		etl.logger.Error("Invalid number of fields in CSV line", "line", line, "field_count", len(fields))
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("expected 5 fields, got %d", len(fields))
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(fields[0]), 10, 64)
	if err != nil {
		etl.logger.Error("Error parsing timestamp", "value", fields[0], "error", err)
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Parse bandwidth_mbps
	bandwidthMbps, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
	if err != nil {
		etl.logger.Error("Error parsing bandwidth_mbps", "value", fields[2], "error", err)
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("invalid bandwidth_mbps: %w", err)
	}

	// Parse latency_ms
	latencyMs, err := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
	if err != nil {
		etl.logger.Error("Error parsing latency_ms", "value", fields[3], "error", err)
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("invalid latency_ms: %w", err)
	}

	// Parse packet_errors
	packetErrors, err := strconv.Atoi(strings.TrimSpace(fields[4]))
	if err != nil {
		etl.logger.Error("Error parsing packet_errors", "value", fields[4], "error", err)
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("invalid packet_errors: %w", err)
	}

	switchID := strings.TrimSpace(fields[1])

	return switchID,
		timestamp,
		telemetrics.MetricRecord{

			BandwidthMbps: bandwidthMbps,
			LatencyMs:     latencyMs,
			PacketErrors:  packetErrors,
		}, nil
}
