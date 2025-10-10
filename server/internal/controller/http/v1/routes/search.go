package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

const (
	Search = "/search"
)

type ISearchService interface {
	Search(ctx *gin.Context)
}

func RegisterSearchRoutes(g *gin.RouterGroup, service ISearchService, securityHandler *jwt.GinJWTMiddleware) {
	g.GET(Search, service.Search)
}
