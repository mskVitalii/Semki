package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/controller/http/v1/routes"
	"time"
)

// statusService - dependent services
type statusService struct {
	repo mongo.IStatusRepository
}

func NewStatusService(repo mongo.IStatusRepository) routes.IStatusService {
	return &statusService{repo: repo}
}

// HealthCheck godoc
//
//	@Summary		Perform health check
//	@Description	Check if the service is healthy
//	@Tags			service
//	@Produce		json
//	@Success		200
//	@Router			/api/v1/healthcheck [get]
func (s statusService) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	if err := s.repo.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message":   "Mongo unavailable",
			"error":     err.Error(),
			"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "OK",
		"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
	})
}

// Metrics godoc
//
//	@Summary		Prometheus logs
//	@Description	Logs that are collected by prometheus and visualized in grafana
//	@Tags			service
//	@Success		200
//	@Router			/metrics [get]
func (s statusService) Metrics(_ *gin.Context) {
	// already implemented. Func is used to generate swagger
}
