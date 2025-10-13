package rateLimit

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

func RedisRateLimit(rdb *redis.Client, limit int, window time.Duration, endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		userIP := c.ClientIP()
		key := fmt.Sprintf("rate:%s:%s", userIP, endpoint)

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			c.Abort()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
