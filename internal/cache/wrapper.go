package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Client RedisClient
}

func GetOrSet[T any](c *Cache, ctx context.Context, key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	var empty T

	val, clientErr := c.Client.Get(ctx, key).Bytes()

	if clientErr == nil {
		var dto T
		if convertErr := json.Unmarshal(val, &dto); convertErr == nil {
			return dto, nil
		} else {
			return empty, convertErr
		}
	}

	if clientErr != redis.Nil {
		return empty, clientErr
	}

	dto, storageErr := fetch()
	if storageErr != nil {
		return empty, storageErr
	}

	data, err := json.Marshal(dto)
	if err == nil {
		_ = c.Client.Set(ctx, key, data, ttl).Err()
	}

	return dto, nil
}
