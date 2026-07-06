package counter

import (
	"context"
	"fmt"
	"time"

	"github.com/svgstat/svgstat/internal/cache"
	"github.com/svgstat/svgstat/internal/project"
)

type Counter struct {
	cache     *cache.Cache
	projectRepo project.Repository
}

func New(cache *cache.Cache, projectRepo project.Repository) *Counter {
	return &Counter{
		cache:     cache,
		projectRepo: projectRepo,
	}
}

func (c *Counter) Increment(ctx context.Context, projectID, counterName string) (int64, error) {
	key := cache.BuildKey("project", projectID, "counter", counterName)
	value, err := c.cache.Increment(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}
	return value, nil
}

func (c *Counter) Get(ctx context.Context, projectID, counterName string) (int64, error) {
	key := cache.BuildKey("project", projectID, "counter", counterName)
	valueStr, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get counter: %w", err)
	}

	var value int64
	_, err = fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return 0, fmt.Errorf("failed to parse counter value: %w", err)
	}

	return value, nil
}

func (c *Counter) Set(ctx context.Context, projectID, counterName string, value int64) error {
	key := cache.BuildKey("project", projectID, "counter", counterName)
	err := c.cache.Set(ctx, key, fmt.Sprintf("%d", value), 0)
	if err != nil {
		return fmt.Errorf("failed to set counter: %w", err)
	}
	return nil
}

func (c *Counter) IncrementWithAnalytics(ctx context.Context, projectID, counterName string) (int64, error) {
	key := cache.BuildKey("project", projectID, "counter", counterName)
	
	pipe := c.cache.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	
	date := time.Now().Format("2006-01-02")
	historyKey := cache.BuildKey("project", projectID, "counter", counterName, "history", date)
	pipe.Incr(ctx, historyKey)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter with analytics: %w", err)
	}
	
	return incrCmd.Val(), nil
}
