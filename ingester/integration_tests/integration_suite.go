package integration_tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	ingesterBaseURL = "http://localhost:8080"
	maxRetries      = 30
	retryDelay      = 2 * time.Second
)

type IntegrationTestSuite struct {
	suite.Suite
	dockerComposeCmd *exec.Cmd
	ctx              context.Context
	cancel           context.CancelFunc
}

// SetupSuite runs once before all tests in the suite
func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Start docker-compose
	s.T().Log("Starting docker-compose services...")
	s.dockerComposeCmd = exec.CommandContext(s.ctx, "docker-compose", "up", "--build")

	// Use relative path - go up two directories from integration_tests to project root
	projectRoot := "../.."
	s.dockerComposeCmd.Dir = projectRoot
	s.dockerComposeCmd.Stdout = os.Stdout
	s.dockerComposeCmd.Stderr = os.Stderr

	err := s.dockerComposeCmd.Start()
	s.Require().NoError(err, "Failed to start docker-compose")

	// time.Sleep(30 * time.Second)

	// Wait for the ingester service to be healthy
	s.T().Log("Waiting for ingester service to be ready...")
	s.waitForService(ingesterBaseURL + "/health")
}

// TearDownSuite runs once after all tests in the suite
func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("Stopping docker-compose services...")

	// Stop docker-compose gracefully
	s.cancel()

	// Run docker-compose down to clean up
	downCmd := exec.Command("docker-compose", "down")
	downCmd.Dir = "../.."
	downCmd.Stdout = os.Stdout
	downCmd.Stderr = os.Stderr

	if err := downCmd.Run(); err != nil {
		s.T().Logf("Warning: Failed to run docker-compose down: %v", err)
	}

	// Wait for the docker-compose process to finish
	if s.dockerComposeCmd != nil && s.dockerComposeCmd.Process != nil {
		_ = s.dockerComposeCmd.Wait()
	}

	s.T().Log("Docker-compose services stopped")
}

// waitForService waits for a service to become available
func (s *IntegrationTestSuite) waitForService(url string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			s.T().Logf("Service at %s is ready", url)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}

		s.T().Logf("Waiting for service at %s (attempt %d/%d)...", url, i+1, maxRetries)
		time.Sleep(retryDelay)
	}

	s.Require().Fail(fmt.Sprintf("Service at %s did not become ready after %d attempts", url, maxRetries))
}
