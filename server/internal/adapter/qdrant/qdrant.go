package qdrant

import (
	"semki/internal/utils/config"
	"semki/pkg/clients"
	"time"
)

func SetupQdrant(cfg *config.QdrantConfig) (*clients.QdrantClient, error) {
	qdrantConfig := clients.QdrantConfig{
		Host:    cfg.Host,
		Port:    cfg.GrpcPort,
		UseTLS:  false,
		Timeout: 10 * time.Second,
	}

	return clients.ConnectToQdrant(qdrantConfig)
}
