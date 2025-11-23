package redis

import (
	"context"
	"fmt"

	"url-shortener-wb/internal/config"

	repo "url-shortener-wb/internal/repository"

	wbfredis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type RedisCache struct {
	client  *wbfredis.Client
	retries retry.Strategy
}

func NewRedisCache(cfg *config.Config, retries retry.Strategy) *RedisCache {
	client := wbfredis.New(cfg.RedisAddr(), cfg.Redis.Pass, cfg.Redis.DB)
	return &RedisCache{
		client:  client,
		retries: retries,
	}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	value, err := c.client.GetWithRetry(ctx, c.retries, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return "", fmt.Errorf("%w: key %s not found", repo.ErrCacheMiss, key)
		}
		return "", fmt.Errorf("failed to get from cache: %w", err)
	}
	return value, nil
}

func (c *RedisCache) Set(ctx context.Context, key, value string) error {
	err := c.client.SetWithRetry(ctx, c.retries, key, value)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	var count int64
	err := retry.DoContext(ctx, c.retries, func() error {
		var err error
		count, err = c.client.Exists(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("redis exists failed: %w", err)
		}
		return nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}
	return count > 0, nil
}
