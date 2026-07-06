package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/svgstat/svgstat/internal/config"
)

type Cache struct {
	client *redis.Client
}

func New(cfg *config.Config) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Info().Msg("Connected to Redis successfully")
	return &Cache{client: client}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *Cache) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

func (c *Cache) HashIncrement(ctx context.Context, key, field string) (int64, error) {
	return c.client.HIncrBy(ctx, key, field, 1).Result()
}

func (c *Cache) SetAdd(ctx context.Context, key, member string) (int64, error) {
	return c.client.SAdd(ctx, key, member).Result()
}

func (c *Cache) SetCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, key).Result()
}

func (c *Cache) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

func (c *Cache) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

func (c *Cache) GetClient() *redis.Client {
	return c.client
}

func BuildKey(parts ...string) string {
	return "svgstat:" + join(parts, ":")
}

func join(parts []string, sep string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += sep
		}
		result += part
	}
	return result
}
