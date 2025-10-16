package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"semki/internal/model"
	"semki/internal/utils/config"
	"semki/internal/utils/crypto"
	"semki/pkg/clients"
)

type IMongoRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	GetUsersByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id primitive.ObjectID, user model.User) error
	DeleteUser(ctx context.Context, id primitive.ObjectID) error
	RestoreUser(ctx context.Context, id primitive.ObjectID) error

	CreateOrganization(ctx context.Context, organization model.Organization) error
	GetOrganizationByID(ctx context.Context, id primitive.ObjectID) (*model.Organization, error)
	GetOrganizationByTitle(ctx context.Context, email string) (*model.Organization, error)
	UpdateOrganization(ctx context.Context, id primitive.ObjectID, organization model.Organization) error
	DeleteOrganization(ctx context.Context, id primitive.ObjectID) error
}

type repository struct {
	config *config.Config
	db     *clients.MongoDb
}

func New(cfg *config.Config, db *clients.MongoDb) IMongoRepository {
	return &repository{cfg, db}
}

//region Users

func (r repository) GetUsersByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Users)

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

func (r repository) CreateUser(ctx context.Context, user *model.User) error {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Users)
	encryptedUser, err := crypto.EncryptUserFields(*user, r.config.CryptoKey)
	if err != nil {
		return err
	}
	_, err = coll.InsertOne(ctx, encryptedUser)
	return err
}

func (r repository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Users)
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
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Users)
	encryptedUser, err := crypto.EncryptUserFields(user, r.config.CryptoKey)
	_, err = coll.ReplaceOne(ctx, bson.M{"_id": id}, bson.M{"$set": encryptedUser})
	return err
}

// DeleteUser performs soft-delete by changing status to "deleted"
func (r repository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Users)

	update := bson.M{"$set": bson.M{"status": model.UserStatuses.DELETED}}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// RestoreUser restores user by changing status to "active"
func (r repository) RestoreUser(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Users)

	update := bson.M{"$set": bson.M{"status": model.UserStatuses.ACTIVE}}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Users)
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

//region Organization

func (r repository) CreateOrganization(ctx context.Context, organization model.Organization) error {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Organizations)
	_, err := coll.InsertOne(ctx, organization)
	return err
}

func (r repository) GetOrganizationByID(ctx context.Context, id primitive.ObjectID) (*model.Organization, error) {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Organizations)
	var organization model.Organization
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&organization)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &organization, nil
}

func (r repository) GetOrganizationByTitle(ctx context.Context, title string) (*model.Organization, error) {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Organizations)
	var organization model.Organization
	err := coll.FindOne(ctx, bson.M{"title": title}).Decode(&organization)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &organization, nil
}

func (r repository) UpdateOrganization(ctx context.Context, id primitive.ObjectID, organization model.Organization) error {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Organizations)
	_, err := coll.ReplaceOne(ctx, bson.M{"_id": id}, organization)
	return err
}

func (r repository) DeleteOrganization(ctx context.Context, id primitive.ObjectID) error {
	coll := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Organizations)
	usersColl := r.db.Client.Database(r.db.Database).Collection(r.db.Collections.Users)

	_, err := usersColl.DeleteMany(ctx, bson.M{"organization_id": id})
	if err != nil {
		return err
	}

	_, err = coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

//endregion
