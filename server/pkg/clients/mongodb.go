package clients

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

// region Collections

type MongoCollectionsNamesType struct {
	Organizations string
	Teams         string
	Levels        string
	Users         string
}

var MongoCollectionsNames = MongoCollectionsNamesType{
	Organizations: "organizations",
	Teams:         "teams",
	Levels:        "levels",
	Users:         "users",
}

// endregion

type MongoDb struct {
	Database    string
	Collections MongoCollectionsNamesType
	Client      *mongo.Client
}

// ConnectToMongoDb connects to a running MongoDB instance
func ConnectToMongoDb(ctx context.Context, user, pass, host, database string, port int) (*MongoDb, error) {
	credential := options.Credential{
		AuthSource: database,
		Username:   user,
		Password:   pass,
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		uri(user, pass, host, database, port),
	).SetAuth(credential).SetMonitor(otelmongo.NewMonitor()))

	if err != nil {
		return nil, errors.Wrap(err, "failed to create MongoDB client")
	}

	// test
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "failed to ping MongoDB server")
	}
	return &MongoDb{
		Database:    database,
		Client:      client,
		Collections: MongoCollectionsNames,
	}, nil
}

// uri generates uri string for connecting to MongoDB.
func uri(user, pass, host, database string, port int) string {
	const format = "mongodb://%s:%s@%s:%d/%s"
	return fmt.Sprintf(format, user, pass, host, port, database)
}
