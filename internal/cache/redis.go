package cache

import (
	"context"
	"time"

	"github.com/maisiq/go-ugc-service/pkg/config"
	"github.com/redis/go-redis/v9"
)

//go:generate minimock -i RedisClient -o ./mocks/ -s "_mock.go"
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Close() error
}

func NewClient(cfg *config.CacheConfig) RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: "",
		DB:       0,
	})
	return rdb
}
