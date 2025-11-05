package dao

import (
	"context"
	"encoding/json"
	"fmt"
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

// AddKey saves a MetricRecord to Redis with the given key
func (dao *DAOMetrics) AddKey(ctx context.Context, key string, record telemetrics.MetricRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return dao.redisClient.Set(ctx, key, data, dao.ttl).Err()
}

// GetAll retrieves all metrics from Redis and returns them as a slice of maps
// Each map contains a single key-value pair where the key is the Redis key
// and the value is the MetricRecord
func (dao *DAOMetrics) GetAll(ctx context.Context) ([]map[string]telemetrics.MetricRecord, error) {
	// Use SCAN instead of KEYS to avoid blocking Redis
	var keys []string
	var cursor uint64
	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = dao.redisClient.Scan(ctx, cursor, "*", 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, scanKeys...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return []map[string]telemetrics.MetricRecord{}, nil
	}

	// Use pipeline to fetch all values in batch
	pipe := dao.redisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Continue even if some keys fail
	}

	// Pre-allocate result slice
	result := make([]map[string]telemetrics.MetricRecord, 0, len(keys))

	for i, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			// Skip keys that don't exist or have errors
			continue
		}

		// Parse the JSON data into MetricRecord
		var record telemetrics.MetricRecord
		if err := json.Unmarshal([]byte(data), &record); err != nil {
			fmt.Printf("Error parsing MetricRecord for key %s: %v\n", keys[i], err)
			continue
		}

		// Add to result as a map with single key-value pair
		result = append(result, map[string]telemetrics.MetricRecord{
			keys[i]: record,
		})
	}

	return result, nil
}

// GetMetric retrieves a specific metric value for a given key from Redis
// Returns the metric value as interface{}, or an error if key or metric doesn't exist
func (dao *DAOMetrics) GetMetric(ctx context.Context, key string, metric string) (interface{}, error) {
	// Get the value for this key
	data, err := dao.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("key does not exist: %s", key)
	}

	// Unmarshal into a map to access individual fields
	var metricMap map[string]interface{}
	if err := json.Unmarshal([]byte(data), &metricMap); err != nil {
		return nil, fmt.Errorf("error parsing data for key %s: %w", key, err)
	}

	// Check if the metric exists in the map
	value, exists := metricMap[metric]
	if !exists {
		return nil, fmt.Errorf("metric '%s' does not exist in key '%s'", metric, key)
	}

	return value, nil
}
