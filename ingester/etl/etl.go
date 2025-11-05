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
	dao      *dao.DAOMetrics
	interval time.Duration
}

func NewETL(dao *dao.DAOMetrics, interval time.Duration) *ETL {
	return &ETL{
		dao:      dao,
		interval: interval,
	}
}

func (etl *ETL) Run() error {
	for {
		if err := etl.updateMetrics(); err != nil {
			fmt.Printf("Error updating metrics: %v\n", err)
		}

		// Sleep until the next interval
		time.Sleep(etl.interval)
	}
}

func (etl *ETL) updateMetrics() error {
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
		etl.writeMetricsLineByLine(resp.Body)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (etl *ETL) writeMetricsLineByLine(respBody io.ReadCloser) error {
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
		switchID, record, err := etl.parseCSVLine(line)
		if err != nil {
			fmt.Printf("Error parsing line %d: %v\n", lineNumber, err)
			continue
		}

		// Store the metric using the DAO
		if err := etl.dao.AddKey(ctx, switchID, record); err != nil {
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
func (etl *ETL) parseCSVLine(line string) (string, telemetrics.MetricRecord, error) {
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
