package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"semki/pkg/clients"
)

type IStatusRepository interface {
	HealthCheck(ctx context.Context) error
}

type statusRepository struct {
	client *clients.MongoDb
}

func NewStatusRepository(client *clients.MongoDb) IStatusRepository {
	return &statusRepository{client}
}

// region Status

func (r *statusRepository) HealthCheck(ctx context.Context) error {
	return r.client.Client.Ping(ctx, readpref.Primary())
}

// endregion
