package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"semki/internal/utils/rateLimit"
	"semki/pkg/lib"
	"time"
)

const (
	chat = "/chat"
)

type IChatService interface {
	CreateChat(c *gin.Context)
	GetChat(c *gin.Context)
	GetUserHistory(c *gin.Context)
}

func RegisterChatRoutes(g *gin.RouterGroup, chatService IChatService, securityHandler gin.HandlerFunc, rds *redis.Client) {
	g.POST(chat, rateLimit.RedisRateLimit(rds, 10, time.Minute, chat), securityHandler, chatService.CreateChat)
	g.OPTIONS(chat+"/:id", lib.Preflight)
	g.GET(chat+"/:id", securityHandler, chatService.GetChat)
	g.OPTIONS(chat+"/history", lib.Preflight)
	g.GET(chat+"/history", securityHandler, chatService.GetUserHistory)
}
