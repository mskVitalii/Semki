package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"semki/internal/model"
	"semki/internal/utils/config"
	"semki/internal/utils/crypto"
	"semki/pkg/clients"
	"time"
)

type IRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	GetUsersByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUsersByOrganization(ctx context.Context, orgID primitive.ObjectID, search string, page, limit int) ([]*model.User, int64, error)
	UpdateUser(ctx context.Context, id primitive.ObjectID, user model.User) error
	DeleteUser(ctx context.Context, id primitive.ObjectID) error
	RestoreUser(ctx context.Context, id primitive.ObjectID) error

	CreateOrganization(ctx context.Context, organization model.Organization) error
	GetOrganizationByID(ctx context.Context, id primitive.ObjectID) (*model.Organization, error)
	GetOrganizationByTitle(ctx context.Context, email string) (*model.Organization, error)
	UpdateOrganization(ctx context.Context, id primitive.ObjectID, organization model.Organization) error
	DeleteOrganization(ctx context.Context, id primitive.ObjectID) error
	PatchOrganization(ctx context.Context, orgID primitive.ObjectID, updates bson.M) error
	AddLevel(ctx context.Context, orgID primitive.ObjectID, level model.Level) error
	UpdateLevel(ctx context.Context, orgID primitive.ObjectID, levelID primitive.ObjectID, updates bson.M) error
	DeleteLevel(ctx context.Context, orgID primitive.ObjectID, levelID primitive.ObjectID) error
	AddTeam(ctx context.Context, orgID primitive.ObjectID, team model.Team) error
	UpdateTeam(ctx context.Context, orgID primitive.ObjectID, teamID primitive.ObjectID, updates bson.M) error
	DeleteTeam(ctx context.Context, orgID primitive.ObjectID, teamID primitive.ObjectID) error
	AddLocation(ctx context.Context, orgID primitive.ObjectID, location model.Location) error
	DeleteLocation(ctx context.Context, orgID primitive.ObjectID, locationID primitive.ObjectID) error

	CreateChat(ctx context.Context, chat *model.Chat) error
	GetChatByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*model.Chat, error)
	GetChatsByUserIDWithCursor(ctx context.Context, userID primitive.ObjectID, cursor string, limit int) ([]model.Chat, string, error)
	DeleteChat(ctx context.Context, id primitive.ObjectID) error
	AddChatMessages(ctx context.Context, chatId primitive.ObjectID, messages []model.Message) error

	HealthCheck(ctx context.Context) error
}

type repository struct {
	config *config.Config
	client *clients.MongoDb
}

func New(cfg *config.Config, client *clients.MongoDb) IRepository {
	return &repository{cfg, client}
}

//region Users

func (r *repository) GetUsersByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
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

func (r *repository) CreateUser(ctx context.Context, user *model.User) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	encryptedUser, err := crypto.EncryptUserFields(*user, r.config.CryptoKey)
	if err != nil {
		return err
	}
	_, err = coll.InsertOne(ctx, encryptedUser)
	return err
}

func (r *repository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
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

func (r *repository) UpdateUser(ctx context.Context, id primitive.ObjectID, user model.User) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)
	encryptedUser, err := crypto.EncryptUserFields(user, r.config.CryptoKey)
	_, err = coll.ReplaceOne(ctx, bson.M{"_id": id}, bson.M{"$set": encryptedUser})
	return err
}

// DeleteUser performs soft-delete by changing status to "deleted"
func (r *repository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)

	update := bson.M{"$set": bson.M{"status": model.UserStatuses.DELETED}}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// RestoreUser restores user by changing status to "active"
func (r *repository) RestoreUser(ctx context.Context, id primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)

	update := bson.M{"$set": bson.M{"status": model.UserStatuses.ACTIVE}}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
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

func (r *repository) GetUsersByOrganization(ctx context.Context, orgID primitive.ObjectID, search string, page, limit int) ([]*model.User, int64, error) {
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

//region Organizations

func (r *repository) CreateOrganization(ctx context.Context, organization model.Organization) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
	_, err := coll.InsertOne(ctx, organization)
	return err
}

func (r *repository) GetOrganizationByID(ctx context.Context, id primitive.ObjectID) (*model.Organization, error) {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
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

func (r *repository) GetOrganizationByTitle(ctx context.Context, title string) (*model.Organization, error) {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
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

func (r *repository) UpdateOrganization(ctx context.Context, id primitive.ObjectID, organization model.Organization) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
	_, err := coll.ReplaceOne(ctx, bson.M{"_id": id}, organization)
	return err
}

func (r *repository) DeleteOrganization(ctx context.Context, id primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
	usersColl := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)

	_, err := usersColl.DeleteMany(ctx, bson.M{"organization_id": id})
	if err != nil {
		return err
	}

	_, err = coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// PatchOrganization updates only provided fields
func (r *repository) PatchOrganization(ctx context.Context, orgID primitive.ObjectID, updates bson.M) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{"$set": updates}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Team methods

func (r *repository) AddTeam(ctx context.Context, orgID primitive.ObjectID, team model.Team) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$push": bson.M{"semantic.teams": team},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *repository) UpdateTeam(ctx context.Context, orgID primitive.ObjectID, teamID primitive.ObjectID, updates bson.M) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	// Build update document for positional operator
	setUpdates := bson.M{}
	for key, value := range updates {
		setUpdates["semantic.teams.$."+key] = value
	}

	filter := bson.M{
		"_id":                orgID,
		"semantic.teams._id": teamID,
	}
	update := bson.M{"$set": setUpdates}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *repository) DeleteTeam(ctx context.Context, orgID primitive.ObjectID, teamID primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$pull": bson.M{
			"semantic.teams": bson.M{"_id": teamID},
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Level methods

func (r *repository) AddLevel(ctx context.Context, orgID primitive.ObjectID, level model.Level) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$push": bson.M{"semantic.levels": level},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *repository) UpdateLevel(ctx context.Context, orgID primitive.ObjectID, levelID primitive.ObjectID, updates bson.M) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	setUpdates := bson.M{}
	for key, value := range updates {
		setUpdates["semantic.levels.$."+key] = value
	}

	filter := bson.M{
		"_id":                 orgID,
		"semantic.levels._id": levelID,
	}
	update := bson.M{"$set": setUpdates}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *repository) DeleteLevel(ctx context.Context, orgID primitive.ObjectID, levelID primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$pull": bson.M{
			"semantic.levels": bson.M{"_id": levelID},
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Location methods

func (r *repository) AddLocation(ctx context.Context, orgID primitive.ObjectID, location model.Location) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$push": bson.M{"semantic.locations": location},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *repository) DeleteLocation(ctx context.Context, orgID primitive.ObjectID, locationID primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$pull": bson.M{
			"semantic.locations": bson.M{"_id": locationID},
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

//endregion

// region Chats

func (r *repository) CreateChat(ctx context.Context, chat *model.Chat) error {
	chat.ID = primitive.NewObjectID()
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()

	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Chats)
	_, err := coll.InsertOne(ctx, chat)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetChatByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*model.Chat, error) {
	var chat model.Chat

	filter := bson.M{
		"_id":    id,
		"userId": userID,
	}

	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Chats)
	err := coll.FindOne(ctx, filter).Decode(&chat)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &chat, nil
}

func (r *repository) GetChatsByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.Chat, error) {
	var chats []*model.Chat

	filter := bson.M{"userId": userID}

	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Chats)
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *repository) GetChatsByUserIDWithCursor(ctx context.Context, userID primitive.ObjectID, cursor string, limit int) ([]model.Chat, string, error) {
	filter := bson.M{"user_id": userID}

	if cursor != "" {
		cursorObjectID, err := primitive.ObjectIDFromHex(cursor)
		if err != nil {
			return nil, "", err
		}
		filter["_id"] = bson.M{"$lt": cursorObjectID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(int64(limit + 1))

	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Chats)

	mongoCursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer mongoCursor.Close(ctx)

	var chats []model.Chat
	if err := mongoCursor.All(ctx, &chats); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(chats) > limit {
		nextCursor = chats[limit].ID.Hex()
		chats = chats[:limit]
	}

	return chats, nextCursor, nil
}

func (r *repository) DeleteChat(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}

	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Chats)
	_, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) AddChatMessages(ctx context.Context, chatID primitive.ObjectID, messages []model.Message) error {
	filter := bson.M{"_id": chatID}
	update := bson.M{
		"$push": bson.M{
			"messages": bson.M{"$each": messages},
		},
	}
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Chats)
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("chat not found")
	}

	return nil
}

// endregion

// region Status

func (r *repository) HealthCheck(ctx context.Context) error {
	return r.client.Client.Ping(ctx, readpref.Primary())
}

// endregion
