package database

import (
	"CloudKey/internal/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis 建立 Redis 连接
func ConnectRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return rdb, nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis(rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	return rdb.Close()
}
