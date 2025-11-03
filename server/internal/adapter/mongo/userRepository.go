package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"semki/internal/model"
	"semki/internal/utils/config"
	"semki/internal/utils/crypto"
	"semki/pkg/clients"
)

type IUserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	GetUsersByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUsersByOrganization(ctx context.Context, orgID primitive.ObjectID, search string, page, limit int) ([]*model.User, int64, error)
	UpdateUser(ctx context.Context, id primitive.ObjectID, user model.User) error
	PatchUser(ctx context.Context, id primitive.ObjectID, data bson.M) error
	DeleteUser(ctx context.Context, id primitive.ObjectID) error
	RestoreUser(ctx context.Context, id primitive.ObjectID) error
}

type userRepository struct {
	config *config.Config
	client *clients.MongoDb
}

func NewUserRepository(cfg *config.Config, client *clients.MongoDb) IUserRepository {
	return &userRepository{cfg, client}
}

//region Users

func (r *userRepository) GetUsersByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)

	filter := bson.M{"_id": bson.M{"$in": ids}}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		decryptedUser, err := crypto.DecryptUserFields(user, r.config.CryptoKey)
		if err != nil {
			return nil, err
		}
		users = append(users, decryptedUser)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	encryptedUser, err := crypto.EncryptUserFields(*user, r.config.CryptoKey)
	if err != nil {
		return err
	}
	_, err = coll.InsertOne(ctx, encryptedUser)
	return err
}

func (r *userRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
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

func (r *userRepository) UpdateUser(ctx context.Context, id primitive.ObjectID, user model.User) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	encryptedUser, err := crypto.EncryptUserFields(user, r.config.CryptoKey)
	_, err = coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": encryptedUser})
	return err
}

func (r *userRepository) PatchUser(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

// DeleteUser performs soft-delete by changing status to "deleted"
func (r *userRepository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)

	update := bson.M{"$set": bson.M{"status": model.UserStatuses.DELETED}}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// RestoreUser restores user by changing status to "active"
func (r *userRepository) RestoreUser(ctx context.Context, id primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)

	update := bson.M{"$set": bson.M{"status": model.UserStatuses.ACTIVE}}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
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

func (r *userRepository) GetUsersByOrganization(ctx context.Context, orgID primitive.ObjectID, search string, page, limit int) ([]*model.User, int64, error) {
	filter := bson.M{"organizationId": orgID}
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	totalCount, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"name": 1})

	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}

//endregion
