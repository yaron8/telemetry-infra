package bootstrap

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/yaron8/telemetry-infra/ingester/config"
	"github.com/yaron8/telemetry-infra/ingester/dao"
	"github.com/yaron8/telemetry-infra/ingester/etl"
	"github.com/yaron8/telemetry-infra/ingester/service"
	"github.com/yaron8/telemetry-infra/telemetrics"
)

type Bootstrap struct {
	config         *config.Config
	allowedMetrics map[string]bool
	apiServer      *service.APIServer
	daoMetrics     *dao.DAOMetrics
}

func NewBootstrap() (*Bootstrap, error) {
	// Load configuration
	cfg := config.NewConfig()

	allowedMetrics := map[string]bool{}
	for _, metric := range telemetrics.GetCSVHeader() {
		if metric != "switch_id" {
			allowedMetrics[metric] = true
		}
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
		Protocol: 2,
	})

	daoMetrics := dao.NewDAOMetrics(redisClient, cfg.Redis.TTL)

	return &Bootstrap{
		config:         cfg,
		allowedMetrics: allowedMetrics,
		apiServer: service.NewAPIServer(
			cfg,
			daoMetrics,
		),
		daoMetrics: daoMetrics,
	}, nil
}

func (b *Bootstrap) Start() error {
	etl := etl.NewETL(
		b.daoMetrics,
		b.config.ETL.Interval,
		b.config.ETL.GeneratorURL,
	)

	go func() {
		etl.Run()
	}()

	return b.apiServer.Start()
}
