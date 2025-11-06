package integration_tests

import (
	"encoding/json"
	"io"
	"net/http"
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

	resp, err := client.Get(ingesterBaseURL + "/health")
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

// TestListMetricsEndpoint tests the /telemetry/ListMetrics endpoint
func (s *IntegrationTestSuite) TestListMetricsEndpoint_StatusOK() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(ingesterBaseURL + "/telemetry/ListMetrics")
	s.Require().NoError(err, "Failed to make request to telemetry/ListMetrics endpoint")
	defer resp.Body.Close()

	// Assert status code is 200
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "Expected status code 200")
}

// MetricData represents the structure of a single metric entry
type MetricData struct {
	BandwidthMbps float64 `json:"bandwidth_mbps"`
	LatencyMs     float64 `json:"latency_ms"`
	PacketErrors  int     `json:"packet_errors"`
}

// TestListMetricsEndpoint_CheckSw5Exists tests the /telemetry/ListMetrics endpoint and checks if sw5 exists
func (s *IntegrationTestSuite) TestListMetricsEndpoint_CheckSwitchExists() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(ingesterBaseURL + "/telemetry/ListMetrics")
	s.Require().NoError(err, "Failed to make request to telemetry/ListMetrics endpoint")
	defer resp.Body.Close()

	// Assert status code is 200
	s.Require().Equal(http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err, "Failed to read response body")

	// Parse JSON response
	var metrics []map[string]interface{}
	err = json.Unmarshal(bodyBytes, &metrics)
	s.Require().NoError(err, "Failed to parse JSON response")

	// Check if sw5 exists in any of the metrics
	sw5Found := false
	var sw5Data MetricData
	for _, metric := range metrics {
		if data, exists := metric["sw5"]; exists {
			sw5Found = true
			// Convert the data to JSON and unmarshal into MetricData struct
			dataBytes, err := json.Marshal(data)
			s.Require().NoError(err, "Failed to marshal sw5 data")
			err = json.Unmarshal(dataBytes, &sw5Data)
			s.Require().NoError(err, "Failed to unmarshal sw5 data into MetricData struct")
			break
		}
	}

	assert.True(s.T(), sw5Found, "Expected to find 'sw5' in the metrics response")

	// Validate the fields exist and have correct types
	if sw5Found {
		// bandwidth_mbps should be a float64
		assert.IsType(s.T(), float64(0), sw5Data.BandwidthMbps, "bandwidth_mbps should be float64")
		assert.Greater(s.T(), sw5Data.BandwidthMbps, float64(0), "bandwidth_mbps should be greater than 0")

		// latency_ms should be a float64
		assert.IsType(s.T(), float64(0), sw5Data.LatencyMs, "latency_ms should be float64")
		assert.GreaterOrEqual(s.T(), sw5Data.LatencyMs, float64(0), "latency_ms should be greater than or equal to 0")

		// packet_errors should be an int
		assert.IsType(s.T(), int(0), sw5Data.PacketErrors, "packet_errors should be int")
		assert.GreaterOrEqual(s.T(), sw5Data.PacketErrors, 0, "packet_errors should be greater than or equal to 0")
	}
}

// TestGetMetricEndpoint tests the /telemetry/GetMetric endpoint
func (s *IntegrationTestSuite) TestGetMetricEndpoint_MetricLatencyMs() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(ingesterBaseURL + "/telemetry/GetMetric?switch_id=sw5&metric=latency_ms")
	s.Require().NoError(err, "Failed to make request to /telemetry/GetMetric endpoint")
	defer resp.Body.Close()

	// Assert status code is 200
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err, "Failed to read response body")

	// Parse JSON response as float64
	var latency float64
	err = json.Unmarshal(bodyBytes, &latency)
	s.Require().NoError(err, "Failed to parse response as float64")

	// Assert the value is float64 and >= 0
	assert.IsType(s.T(), float64(0), latency, "Response should be float64")
	assert.GreaterOrEqual(s.T(), latency, float64(0), "latency_ms should be greater than or equal to 0")
}
