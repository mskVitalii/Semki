package mongo

import (
	"context"
	"dwt/internal/model"
	"dwt/internal/utils/config"
	"dwt/internal/utils/crypto"
	"dwt/pkg/clients"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IMongoRepository interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id primitive.ObjectID, user model.User) error
	DeleteUser(ctx context.Context, id primitive.ObjectID) error
}

type repository struct {
	config *config.Config
	client *clients.MongoDb
}

func New(cfg *config.Config, client *clients.MongoDb) IMongoRepository {
	return &repository{cfg, client}
}

//region Users

func (r repository) CreateUser(ctx context.Context, user model.User) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	encryptedUser, err := crypto.EncryptUserFields(user, r.config.CryptoKey)
	if err != nil {
		return err
	}
	_, err = coll.InsertOne(ctx, encryptedUser)
	return err
}

func (r repository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	var user model.User
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	decryptedUser, err := crypto.DecryptUserFields(user, r.config.CryptoKey)
	if err != nil {
		return nil, err
	}
	return decryptedUser, nil
}

func (r repository) UpdateUser(ctx context.Context, id primitive.ObjectID, user model.User) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	encryptedUser, err := crypto.EncryptUserFields(user, r.config.CryptoKey)
	_, err = coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": encryptedUser})
	return err
}

// DeleteUser performs soft-delete by changing status to "deleted"
func (r repository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)

	update := bson.M{"$set": bson.M{"status": model.UserStatuses.DELETED}}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	var user model.User
	err := coll.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	decryptedUser, err := crypto.DecryptUserFields(user, r.config.CryptoKey)
	if err != nil {
		return nil, err
	}
	return decryptedUser, nil
}

//endregion

// region Organization
// TODO: Organization
// endregion
