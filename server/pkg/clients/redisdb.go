package clients

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"semki/internal/utils/config"
)

func ConnectToRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
	})
}
