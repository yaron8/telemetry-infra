package bootstrap

import (
	"github.com/yaron8/telemetry-infra/generator/config"
	"github.com/yaron8/telemetry-infra/generator/metrics"
	"github.com/yaron8/telemetry-infra/generator/service"
	"github.com/yaron8/telemetry-infra/logi"
)

type Bootstrap struct {
	config    *config.Config
	apiServer *service.APIServer
}

func NewBootstrap() (*Bootstrap, error) {
	// Initialize logger
	_, err := logi.NewLog(nil)
	if err != nil {
		return nil, err
	}

	// Load configuration
	cfg := config.NewConfig()

	apiServer := service.NewAPIServer(
		cfg,
		metrics.NewCSVMetrics(cfg.SnapshotTTL),
	)

	return &Bootstrap{
		apiServer: apiServer,
		config:    cfg,
	}, nil
}

func (b *Bootstrap) Start() error {
	logger := logi.GetLogger()
	logger.Info("Bootstrap is starting")
	return b.apiServer.Start()
}
