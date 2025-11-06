package integration_tests

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(IntegrationTestSuite))
}

// TestHealthEndpoint tests the /health endpoint
func (s *IntegrationTestSuite) TestHealthEndpoint() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(generatorBaseURL + "/health")
	s.Require().NoError(err, "Failed to make request to /health endpoint")
	defer resp.Body.Close()

	// Assert status code is 200
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Read and assert response body
	bodyBytes := make([]byte, 100) // Allocate enough space
	n, _ := resp.Body.Read(bodyBytes)
	body := string(bodyBytes[:n])

	assert.Equal(s.T(), "OK", body, "Expected response body to be 'OK'")
}

// TestCountersEndpoint tests the /counters endpoint
// Calls the endpoint for 15 seconds and verifies we get at least one 200 and at least one 304
// Also verifies that the 200 response contains valid CSV with the expected header and data types
func (s *IntegrationTestSuite) TestCountersEndpoint() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	endTime := time.Now().Add(15 * time.Second)
	got200 := false
	got304 := false
	csvValidated := false

	expectedHeader := []string{"timestamp", "switch_id", "bandwidth_mbps", "latency_ms", "packet_errors"}

	for time.Now().Before(endTime) {
		resp, err := client.Get(generatorBaseURL + "/counters")
		if err != nil {
			s.T().Logf("Request failed: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		statusCode := resp.StatusCode

		switch statusCode {
		case http.StatusOK:
			got200 = true
			s.T().Log("Got 200 OK response")

			// Parse and validate CSV
			if !csvValidated {
				reader := csv.NewReader(resp.Body)
				reader.TrimLeadingSpace = true

				// Read all records
				records, err := reader.ReadAll()
				resp.Body.Close()

				if err != nil {
					s.T().Logf("Failed to parse CSV: %v", err)
					break
				}

				// Validate we have at least header + 1 data row
				s.Require().Greater(len(records), 0, "CSV should have at least a header")

				// Validate header (first row)
				header := records[0]
				s.T().Logf("CSV header: %v", header)
				assert.Equal(s.T(), 5, len(header), "CSV header should have exactly 5 fields")
				assert.Equal(s.T(), expectedHeader, header, "CSV header should match expected format")

				// Validate all data rows
				for i := 1; i < len(records); i++ {
					row := records[i]
					s.T().Logf("Validating row %d: %v", i, row)

					// Ensure exactly 5 fields per row
					assert.Equal(s.T(), 5, len(row), "Each CSV row should have exactly 5 fields")

					// Validate timestamp (int64)
					timestamp, err := strconv.ParseInt(row[0], 10, 64)
					assert.NoError(s.T(), err, "timestamp should be a valid int64")
					assert.Greater(s.T(), timestamp, int64(0), "timestamp should be positive")

					// Validate switch_id (string - just check it's not empty)
					assert.NotEmpty(s.T(), row[1], "switch_id should not be empty")

					// Validate bandwidth_mbps (float64)
					bandwidth, err := strconv.ParseFloat(row[2], 64)
					assert.NoError(s.T(), err, "bandwidth_mbps should be a valid float64")
					assert.GreaterOrEqual(s.T(), bandwidth, 0.0, "bandwidth_mbps should be non-negative")

					// Validate latency_ms (float64)
					latency, err := strconv.ParseFloat(row[3], 64)
					assert.NoError(s.T(), err, "latency_ms should be a valid float64")
					assert.GreaterOrEqual(s.T(), latency, 0.0, "latency_ms should be non-negative")

					// Validate packet_errors (int)
					packetErrors, err := strconv.Atoi(row[4])
					assert.NoError(s.T(), err, "packet_errors should be a valid int")
					assert.GreaterOrEqual(s.T(), packetErrors, 0, "packet_errors should be non-negative")
				}

				csvValidated = true
				s.T().Logf("CSV validation complete: %d data rows validated", len(records)-1)
			} else {
				resp.Body.Close()
			}

		case http.StatusNotModified:
			got304 = true
			s.T().Log("Got 304 Not Modified response")
			resp.Body.Close()

		default:
			resp.Body.Close()
		}

		// If we got both, we can exit early
		if got200 && got304 && csvValidated {
			s.T().Log("Got both 200 and 304, and CSV validated, test conditions met")
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Assert that we got at least one 200 and at least one 304
	assert.True(s.T(), got200, "Expected at least one 200 OK response")
	assert.True(s.T(), got304, "Expected at least one 304 Not Modified response")
	assert.True(s.T(), csvValidated, "Expected to validate CSV structure and data types from 200 response")
}
