package metrics

import (
	"encoding/csv"
	"strings"
	"sync"
	"testing"
	"time"
)

const testCacheDuration = 10 * time.Second

// TestGetCSVMetrics_BasicFunctionality tests that GetCSVMetrics returns valid CSV data
func TestGetCSVMetrics_BasicFunctionality(t *testing.T) {
	cm := NewCSVMetrics(testCacheDuration)

	csvData, err := cm.GetCSVMetrics()
	if err != nil {
		t.Fatalf("GetCSVMetrics() returned error: %v", err)
	}

	if csvData == "" {
		t.Fatal("GetCSVMetrics() returned empty string")
	}

	// Verify it's valid CSV by parsing it
	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV data: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("CSV data has no records")
	}
}

// TestGetCSVMetrics_CSVFormat tests the CSV format and structure
func TestGetCSVMetrics_CSVFormat(t *testing.T) {
	cm := NewCSVMetrics(testCacheDuration)

	csvData, err := cm.GetCSVMetrics()
	if err != nil {
		t.Fatalf("GetCSVMetrics() returned error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV data: %v", err)
	}

	// Check header
	expectedHeader := []string{
		"timestamp",
		"switch_ame",
		"bandwidth_mbps",
		"latency_ms",
		"packet_errors",
	}

	if len(records) < 1 {
		t.Fatal("CSV has no header")
	}

	header := records[0]
	if len(header) != len(expectedHeader) {
		t.Fatalf("Header has %d columns, expected %d", len(header), len(expectedHeader))
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			t.Errorf("Header column %d: got %q, want %q", i, header[i], expected)
		}
	}

	// Check that we have exactly numOfDataLines (100) data rows + 1 header row
	expectedRows := numOfDataLines + 1
	if len(records) != expectedRows {
		t.Errorf("Expected %d total rows (1 header + %d data), got %d", expectedRows, numOfDataLines, len(records))
	}

	// Verify data row structure (all rows should have 5 columns)
	for i := 1; i < len(records); i++ {
		if len(records[i]) != 5 {
			t.Errorf("Row %d has %d columns, expected 5", i, len(records[i]))
		}
	}
}

// TestGetCSVMetrics_CacheReturnsIdenticalData tests that cache returns the same data
func TestGetCSVMetrics_CacheReturnsIdenticalData(t *testing.T) {
	cm := NewCSVMetrics(testCacheDuration)

	// First call - should generate new data
	firstCall, err := cm.GetCSVMetrics()
	if err != nil {
		t.Fatalf("First call returned error: %v", err)
	}

	// Second call immediately after - should return cached data
	secondCall, err := cm.GetCSVMetrics()
	if err != nil {
		t.Fatalf("Second call returned error: %v", err)
	}

	if firstCall != secondCall {
		t.Error("Second call within cache duration returned different data (should be cached)")
	}

	// Third call after short delay - should still be cached
	time.Sleep(1 * time.Second)
	thirdCall, err := cm.GetCSVMetrics()
	if err != nil {
		t.Fatalf("Third call returned error: %v", err)
	}

	if firstCall != thirdCall {
		t.Error("Third call within cache duration returned different data (should be cached)")
	}
}

// TestGetCSVMetrics_CacheExpires tests that cache expires after 10 seconds
func TestGetCSVMetrics_CacheExpires(t *testing.T) {
	// This test takes ~11 seconds to run
	if testing.Short() {
		t.Skip("Skipping cache expiration test in short mode")
	}

	cm := NewCSVMetrics(testCacheDuration)

	// First call - should generate new data
	firstCall, err := cm.GetCSVMetrics()
	if err != nil {
		t.Fatalf("First call returned error: %v", err)
	}

	// Wait for cache to expire (10 seconds + buffer)
	t.Log("Waiting for cache to expire (11 seconds)...")
	time.Sleep(11 * time.Second)

	// Call after cache expiration - should generate new data
	secondCall, err := cm.GetCSVMetrics()
	if err != nil {
		t.Fatalf("Second call after cache expiry returned error: %v", err)
	}

	// The data should be different because:
	// 1. Random values will be different
	// 2. Timestamp will be different
	if firstCall == secondCall {
		t.Error("Call after cache expiry returned identical data (should be regenerated)")
	}
}

// TestGetCSVMetrics_ConcurrentAccess tests thread safety with concurrent calls
func TestGetCSVMetrics_ConcurrentAccess(t *testing.T) {
	cm := NewCSVMetrics(testCacheDuration)
	numGoroutines := 50
	var wg sync.WaitGroup

	// Channel to collect results
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Launch multiple goroutines that call GetCSVMetrics concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := cm.GetCSVMetrics()
			if err != nil {
				errors <- err
				return
			}
			results <- data
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent call returned error: %v", err)
	}

	// Collect all results
	var allResults []string
	for data := range results {
		allResults = append(allResults, data)
	}

	if len(allResults) != numGoroutines {
		t.Fatalf("Expected %d results, got %d", numGoroutines, len(allResults))
	}

	// All concurrent calls should return the same cached data
	firstResult := allResults[0]
	for i, result := range allResults {
		if result != firstResult {
			t.Errorf("Result %d differs from first result (cache should be consistent)", i)
		}
	}
}

// TestGetCSVMetrics_DataFormat tests that generated data has valid format
func TestGetCSVMetrics_DataFormat(t *testing.T) {
	cm := NewCSVMetrics(testCacheDuration)

	csvData, err := cm.GetCSVMetrics()
	if err != nil {
		t.Fatalf("GetCSVMetrics() returned error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV data: %v", err)
	}

	// Skip header, check data rows
	for i := 1; i < len(records); i++ {
		row := records[i]

		// Check timestamp is not empty
		if row[0] == "" {
			t.Errorf("Row %d: timestamp is empty", i)
		}

		// Check switch name format (should be sw1, sw2, etc.)
		if !strings.HasPrefix(row[1], "sw") {
			t.Errorf("Row %d: switch name %q doesn't start with 'sw'", i, row[1])
		}

		// Check that numeric fields are not empty
		if row[2] == "" {
			t.Errorf("Row %d: bandwidth_mbps is empty", i)
		}
		if row[3] == "" {
			t.Errorf("Row %d: latency_ms is empty", i)
		}
		if row[4] == "" {
			t.Errorf("Row %d: packet_errors is empty", i)
		}
	}
}

// TestNewCSVMetrics tests the constructor
func TestNewCSVMetrics(t *testing.T) {
	cm := NewCSVMetrics(testCacheDuration)

	if cm == nil {
		t.Fatal("NewCSVMetrics() returned nil")
	}

	// Verify initial cache is empty
	if cm.cachedData != "" {
		t.Error("New CSVMetrics should have empty cachedData")
	}

	// Verify cacheTime is zero value
	if !cm.cacheTime.IsZero() {
		t.Error("New CSVMetrics should have zero cacheTime")
	}

	// Verify cacheDuration is set correctly
	if cm.cacheTTL != testCacheDuration {
		t.Errorf("cacheDuration = %v, want %v", cm.cacheTTL, testCacheDuration)
	}
}

// TestGetCSVMetrics_MultipleInstances tests that different instances have separate caches
func TestGetCSVMetrics_MultipleInstances(t *testing.T) {
	cm1 := NewCSVMetrics(testCacheDuration)
	cm2 := NewCSVMetrics(testCacheDuration)

	data1, err := cm1.GetCSVMetrics()
	if err != nil {
		t.Fatalf("Instance 1 returned error: %v", err)
	}

	// Small delay to ensure different timestamp
	time.Sleep(100 * time.Millisecond)

	data2, err := cm2.GetCSVMetrics()
	if err != nil {
		t.Fatalf("Instance 2 returned error: %v", err)
	}

	// Different instances should potentially have different data due to random generation
	// But at minimum, they should both return valid data
	if data1 == "" || data2 == "" {
		t.Error("One of the instances returned empty data")
	}
}
