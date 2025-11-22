package redis

import (
	"context"
	"url-shortener-wb/internal/config"

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
	return c.client.GetWithRetry(ctx, c.retries, key)
}

func (c *RedisCache) Set(ctx context.Context, key, value string) error {
	return c.client.SetWithRetry(ctx, c.retries, key, value)
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	var count int64
	err := retry.DoContext(ctx, c.retries, func() error {
		var err error
		count, err = c.client.Exists(ctx, key).Result()
		return err
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
