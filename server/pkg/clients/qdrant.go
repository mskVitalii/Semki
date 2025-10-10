package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type QdrantConfig struct {
	Host    string
	Port    int
	APIKey  string
	UseTLS  bool
	Timeout time.Duration
}

type QdrantClient struct {
	client      qdrant.QdrantClient
	conn        *grpc.ClientConn
	Collections qdrant.CollectionsClient
	Points      qdrant.PointsClient
}

func ConnectToQdrant(cfg QdrantConfig) (*QdrantClient, error) {
	// Формируем адрес подключения
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Настройки gRPC подключения
	var opts []grpc.DialOption

	// Настройка TLS/SSL
	if cfg.UseTLS {
		// Для production используйте правильные TLS credentials
		// opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Добавляем API ключ если он есть
	if cfg.APIKey != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&apiKeyAuth{
			apiKey: cfg.APIKey,
		}))
	}

	// Устанавливаем таймаут
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Создаем gRPC подключение
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	// Создаем клиентов для различных сервисов
	client := &QdrantClient{
		conn:        conn,
		Collections: qdrant.NewCollectionsClient(conn),
		Points:      qdrant.NewPointsClient(conn),
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	_, err = client.Collections.List(ctx, &qdrant.ListCollectionsRequest{})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to verify Qdrant connection: %w", err)
	}

	return client, nil
}

// region API KEY AUTH

type apiKeyAuth struct {
	apiKey string
}

func (a *apiKeyAuth) RequireTransportSecurity() bool {
	return false
}

func (a *apiKeyAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"api-key": a.apiKey,
	}, nil
}

// endregion

// region Qdrant Client Methods

func (c *QdrantClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// endregion
