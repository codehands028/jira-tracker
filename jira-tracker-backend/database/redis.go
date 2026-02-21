package database

import (
	"context"
	"fmt"
	"log"

	"jira-tracker-backend/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var ctx = context.Background()

func InitRedis() error {
	cfg := config.GlobalConfig.Redis
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	log.Println("Redis initialized successfully")
	return nil
}
