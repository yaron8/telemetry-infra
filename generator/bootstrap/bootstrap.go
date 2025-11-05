package bootstrap

import (
	"github.com/yaron8/telemetry-infra/generator/config"
	"github.com/yaron8/telemetry-infra/generator/metrics"
	"github.com/yaron8/telemetry-infra/generator/service"
)

type Bootstrap struct {
	config    *config.Config
	apiServer *service.APIServer
}

func NewBootstrap() (*Bootstrap, error) {
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
	return b.apiServer.Start()
}
