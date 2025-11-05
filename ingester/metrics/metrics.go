package metrics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yaron8/telemetry-infra/telemetrics"
)

// Metrics handles telemetry metrics storage and retrieval
type Metrics struct {
	redisClient *redis.Client
	ttl         time.Duration
}

// NewMetrics creates a new Metrics instance with the provided Redis client
func NewMetrics(redisClient *redis.Client, ttl time.Duration) *Metrics {
	return &Metrics{
		redisClient: redisClient,
	}
}

// Store saves a MetricRecord to Redis with the given key
func (m *Metrics) Store(ctx context.Context, key string, record telemetrics.MetricRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return m.redisClient.Set(ctx, key, data, m.ttl).Err()
}
