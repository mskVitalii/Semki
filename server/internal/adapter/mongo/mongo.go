package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"semki/internal/utils/config"
	"semki/pkg/clients"
	"semki/pkg/telemetry"
)

func SetupMongo(cfg *config.MongoConfig) (*clients.MongoDb, error) {
	ctx := context.Background()
	db, err := clients.ConnectToMongoDb(ctx,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Database,
		cfg.Port)
	if err != nil {
		return nil, err
	}

	if err = db.Client.Ping(ctx, nil); err != nil {
		telemetry.Log.Fatal("failed to ping MongoDB server", zap.Error(err))
		return nil, err
	}
	telemetry.Log.Info("Connected to MongoDB")

	if err := CreateUserCollection(db); err != nil {
		telemetry.Log.Fatal("failed to create user collection", zap.Error(err))
		return nil, err
	}

	if err := CreateOrganizationCollection(db); err != nil {
		telemetry.Log.Fatal("failed to create organizations collection", zap.Error(err))
		return nil, err
	}

	return db, nil
}

func CreateUserCollection(db *clients.MongoDb) error {
	ctx := context.Background()
	err := db.Client.Database(db.Database).CreateCollection(ctx, db.Collections.Users)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return err
	}
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}}, // index in ascending order or use -1 for descending order
		Options: options.Index().SetUnique(true)}

	coll := db.Client.Database(db.Database).Collection(db.Collections.Users)
	if _, err = coll.Indexes().CreateOne(ctx, indexModel); err != nil {
		return err
	}

	return nil
}

func CreateOrganizationCollection(db *clients.MongoDb) error {
	ctx := context.Background()
	err := db.Client.Database(db.Database).CreateCollection(ctx, db.Collections.Organizations)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return err
	}

	coll := db.Client.Database(db.Database).Collection(db.Collections.Organizations)

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "title", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "semantic.levels.name", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "semantic.teams.name", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "semantic.locations.name", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	}

	for _, model := range indexes {
		if _, err := coll.Indexes().CreateOne(ctx, model); err != nil {
			return err
		}
	}

	return nil
}
