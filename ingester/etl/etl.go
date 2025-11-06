package etl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yaron8/telemetry-infra/ingester/dao"
	"github.com/yaron8/telemetry-infra/telemetrics"
)

type ETL struct {
	dao          *dao.DAOMetrics
	interval     time.Duration
	generatorURL string
}

func NewETL(dao *dao.DAOMetrics, interval time.Duration, generatorURL string) *ETL {
	return &ETL{
		dao:          dao,
		interval:     interval,
		generatorURL: generatorURL,
	}
}

func (etl *ETL) Run() {
	for {
		if err := etl.updateMetrics(); err != nil {
			fmt.Printf("Error updating metrics: %v\n", err)
		}

		// Sleep until the next interval
		time.Sleep(etl.interval)
	}
}

func (etl *ETL) updateMetrics() error {
	resp, err := http.Get(etl.generatorURL + "/counters")
	if err != nil {
		fmt.Println("error")
		return fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		fmt.Println("304 - skip")
	case http.StatusOK:
		if err := etl.writeMetricsLineByLine(resp.Body); err != nil {
			return fmt.Errorf("failed to write metrics: %w", err)
		}
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (etl *ETL) writeMetricsLineByLine(respBody io.ReadCloser) error {
	fmt.Println("Pulling metrics from service Generator..")

	scanner := bufio.NewScanner(respBody)
	ctx := context.Background()

	// Skip the header line
	if !scanner.Scan() {
		// No data to read
		return nil
	}

	lastTimeUpdated := int64(0)

	lineNumber := 1
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
			fmt.Printf("Error parsing line %d: %v\n", lineNumber, err)
			continue
		}

		// Store the metric using the DAO
		if err := etl.dao.AddMetric(ctx, timestamp, switchID, record); err != nil {
			fmt.Printf("Error storing metric at line %d: %v\n", lineNumber, err)
			continue
		}

		if timestamp > 0 {
			lastTimeUpdated = timestamp
		}
	}

	// Update key in Redis for last update time
	if lastTimeUpdated > 0 {
		etl.dao.SetLastUpdateTime(ctx, lastTimeUpdated)
	} else {
		fmt.Println("Error: No valid timestamp found to update last update time")
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
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("expected 5 fields, got %d", len(fields))
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(fields[0]), 10, 64)
	if err != nil {
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Parse bandwidth_mbps
	bandwidthMbps, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
	if err != nil {
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("invalid bandwidth_mbps: %w", err)
	}

	// Parse latency_ms
	latencyMs, err := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
	if err != nil {
		return "", 0, telemetrics.MetricRecord{}, fmt.Errorf("invalid latency_ms: %w", err)
	}

	// Parse packet_errors
	packetErrors, err := strconv.Atoi(strings.TrimSpace(fields[4]))
	if err != nil {
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
