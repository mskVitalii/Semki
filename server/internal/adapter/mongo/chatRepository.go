package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"semki/internal/model"
	"semki/pkg/clients"
	"time"
)

type IChatRepository interface {
	CreateChat(ctx context.Context, chat *model.Chat) error
	GetChatByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*model.Chat, error)
	GetChatsByUserIDWithCursor(ctx context.Context, userID primitive.ObjectID, cursor string, limit int) ([]model.Chat, string, error)
	DeleteChat(ctx context.Context, id primitive.ObjectID) error
	AddChatMessages(ctx context.Context, chatId primitive.ObjectID, messages []model.Message) error
}

type chatRepository struct {
	client *clients.MongoDb
}

func NewChatRepository(client *clients.MongoDb) IChatRepository {
	return &chatRepository{client}
}

// region Chats

func (r *chatRepository) CreateChat(ctx context.Context, chat *model.Chat) error {
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

func (r *chatRepository) GetChatByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*model.Chat, error) {
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

func (r *chatRepository) GetChatsByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.Chat, error) {
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

func (r *chatRepository) GetChatsByUserIDWithCursor(ctx context.Context, userID primitive.ObjectID, cursor string, limit int) ([]model.Chat, string, error) {
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

func (r *chatRepository) DeleteChat(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}

	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Chats)
	_, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepository) AddChatMessages(ctx context.Context, chatID primitive.ObjectID, messages []model.Message) error {
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
