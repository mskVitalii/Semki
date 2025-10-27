package rateLimit

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"semki/pkg/lib"
	"time"
)

func RedisRateLimit(rdb *redis.Client, limit int, window time.Duration, endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		userIP := c.ClientIP()
		key := fmt.Sprintf("rate:%s:%s", userIP, endpoint)

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			lib.ResponseInternalServerError(c, err, "Internal error")
			c.Abort()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
