package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"semki/internal/utils/rateLimit"
	"semki/pkg/lib"
	"time"
)

const (
	Search = "/search"
)

type ISearchService interface {
	Search(ctx *gin.Context)
}

func RegisterSearchRoutes(g *gin.RouterGroup, service ISearchService, sec gin.HandlerFunc, rds *redis.Client) {
	g.OPTIONS(Search, lib.Preflight)
	g.GET(Search, rateLimit.RedisRateLimit(rds, 10, time.Minute, Search), sec, service.Search)
}
