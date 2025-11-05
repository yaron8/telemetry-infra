package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/yaron8/telemetry-infra/metrics"
)

const numOfDataLines = 100

func main() {
	// Create or open the CSV file
	file, err := os.Create("generator/csv_metrics/metrics.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header with exact format requested (note: "switch_ame" appears to be intentional typo from original)
	header := []string{"timestamp", "switch_ame", "bandwidth_mbps", "latency_ms", "packet_errors"}
	if err := writer.Write(header); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing header: %v\n", err)
		os.Exit(1)
	}

	// Get current timestamp
	currentTime := time.Now().Unix()

	// Generate numOfDataLines lines of data
	for i := 0; i < numOfDataLines; i++ {
		// Generate random metrics data
		metric := metrics.Metric{
			Timestamp:     currentTime,
			SwitchName:    fmt.Sprintf("sw%d", i+1),
			BandwidthMbps: rand.Float64() * 10000, // Random bandwidth up to 10 Gbps
			LatencyMs:     rand.Float64() * 1500,  // Random latency up to 15000ms
			PacketErrors:  rand.Intn(10),          // Random packet errors up to 1000
		}

		row := []string{
			fmt.Sprintf("%d", metric.Timestamp),
			metric.SwitchName,
			fmt.Sprintf("%.2f", metric.BandwidthMbps),
			fmt.Sprintf("%.2f", metric.LatencyMs),
			fmt.Sprintf("%d", metric.PacketErrors),
		}

		if err := writer.Write(row); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing row: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("CSV file generated successfully at metrics.csv")
}
