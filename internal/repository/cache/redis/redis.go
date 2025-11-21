package redis

import (
	"context"

	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type URLCache struct {
	client  *redis.Client
	retries retry.Strategy
}

func NewURLCache(
	client *redis.Client,
	retries retry.Strategy,
) *URLCache {
	return &URLCache{
		client:  client,
		retries: retries,
	}
}

func (c *URLCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.GetWithRetry(ctx, c.retries, key)
}

func (c *URLCache) Set(ctx context.Context, key, value string) error {
	return c.client.SetWithRetry(ctx, c.retries, key, value)
}

// func (c *URLCache) Exists(ctx context.Context, key string) (bool, error) {
// 	var exists bool
// 	err := retry.DoContext(ctx, c.retries, func(ctx context.Context) error {
// 		count, err := c.client.Exists(ctx, key)
// 		if err != nil {
// 			return err
// 		}
// 		exists = count.(int64) > 0
// 		return nil
// 	})

// 	return exists, err
// }
