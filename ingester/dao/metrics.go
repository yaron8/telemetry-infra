package dao

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yaron8/telemetry-infra/telemetrics"
)

// DAOMetrics handles telemetry metrics storage and retrieval
type DAOMetrics struct {
	redisClient *redis.Client
	ttl         time.Duration
}

// NewDAOMetrics creates a new Metrics instance with the provided Redis client
func NewDAOMetrics(redisClient *redis.Client, ttl time.Duration) *DAOMetrics {
	return &DAOMetrics{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

// Store saves a MetricRecord to Redis with the given key
func (dao *DAOMetrics) Store(ctx context.Context, key string, record telemetrics.MetricRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return dao.redisClient.Set(ctx, key, data, dao.ttl).Err()
}
