package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	ReIndex = "/reindex"
)

type IQdrantController interface {
	ReIndexReq(c *gin.Context)
}

func RegisterQdrantRoutes(g *gin.RouterGroup, securityHandler gin.HandlerFunc, service IQdrantController) {
	g.POST(ReIndex, securityHandler, service.ReIndexReq)
}
