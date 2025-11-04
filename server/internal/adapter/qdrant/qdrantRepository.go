package qdrant

import (
	"context"
	"fmt"
	"time"

	"semki/internal/model"
	"semki/internal/utils/config"
	"semki/pkg/clients"

	"github.com/qdrant/go-client/qdrant"
)

const UsersCollection = "users"

type SearchFilters struct {
	Query     string   `json:"query"`
	Teams     []string `json:"teams"`
	Levels    []string `json:"levels"`
	Locations []string `json:"locations"`
	Limit     uint64   `json:"limit"`
}

type VectorSearchResult struct {
	Score  float32
	UserID string
}

type IQdrantRepository interface {
	InitializeCollection(ctx context.Context) error
	IndexUserWithVector(ctx context.Context, user *model.User, vector []float32) error
	UpdateUserWithVector(ctx context.Context, user *model.User, vector []float32) error
	DeleteUser(ctx context.Context, id string) error
	SearchUserByVector(ctx context.Context, vector []float32, filter SearchFilters) ([]VectorSearchResult, error)
}

type repository struct {
	client         *clients.QdrantClient
	collectionName string
	vectorSize     uint64
}

func New(cfg *config.Config, client *clients.QdrantClient) IQdrantRepository {
	repo := &repository{
		client:         client,
		collectionName: UsersCollection,
		vectorSize:     uint64(cfg.Embedder.Dimensions),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := repo.InitializeCollection(ctx); err != nil {
		fmt.Printf("Warning: failed to initialize collection: %v\n", err)
	}

	return repo
}

func (r *repository) InitializeCollection(ctx context.Context) error {
	collections, err := r.client.Collections.List(ctx, &qdrant.ListCollectionsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	exists := false
	for _, c := range collections.Collections {
		if c.Name == r.collectionName {
			exists = true
			break
		}
	}

	if !exists {
		_, err = r.client.Collections.Create(ctx, &qdrant.CreateCollection{
			CollectionName: r.collectionName,
			VectorsConfig: &qdrant.VectorsConfig{
				Config: &qdrant.VectorsConfig_Params{
					Params: &qdrant.VectorParams{
						Size:     r.vectorSize,
						Distance: qdrant.Distance_Cosine,
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}

		if err := r.createPayloadIndexes(ctx); err != nil {
			return fmt.Errorf("failed to create payload indexes: %w", err)
		}
	}

	return nil
}

func (r *repository) createPayloadIndexes(ctx context.Context) error {
	indexes := []struct {
		fieldName string
		fieldType qdrant.FieldType
	}{
		{"user_id", qdrant.FieldType_FieldTypeKeyword},
	}

	for _, idx := range indexes {
		_, err := r.client.Points.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
			CollectionName: r.collectionName,
			FieldName:      idx.fieldName,
			FieldType:      &idx.fieldType,
		})
		if err != nil {
			return fmt.Errorf("failed to create index for %s: %w", idx.fieldName, err)
		}
	}

	return nil
}

func (r *repository) IndexUserWithVector(ctx context.Context, user *model.User, vector []float32) error {
	pointID, err := r.userIDToPointID(user.ID.Hex())
	if err != nil {
		return fmt.Errorf("failed to convert user ID: %w", err)
	}

	payload, err := r.userToPayload(user)
	if err != nil {
		return fmt.Errorf("failed to create payload: %w", err)
	}

	point := &qdrant.PointStruct{
		Id: &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Num{
				Num: pointID,
			},
		},
		Vectors: &qdrant.Vectors{
			VectorsOptions: &qdrant.Vectors_Vector{
				Vector: &qdrant.Vector{Data: vector},
			},
		},
		Payload: payload,
	}

	_, err = r.client.Points.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: r.collectionName,
		Points:         []*qdrant.PointStruct{point},
	})
	if err != nil {
		return fmt.Errorf("failed to index user: %w", err)
	}

	return nil
}

func (r *repository) UpdateUserWithVector(ctx context.Context, user *model.User, vector []float32) error {
	return r.IndexUserWithVector(ctx, user, vector)
}

func (r *repository) DeleteUser(ctx context.Context, id string) error {
	pointID, err := r.userIDToPointID(id)
	if err != nil {
		return fmt.Errorf("failed to convert user ID: %w", err)
	}

	_, err = r.client.Points.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: r.collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: []*qdrant.PointId{
						{PointIdOptions: &qdrant.PointId_Num{Num: pointID}},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *repository) SearchUserByVector(ctx context.Context, vector []float32, filters SearchFilters) ([]VectorSearchResult, error) {
	response, err := r.client.Points.Search(ctx, &qdrant.SearchPoints{
		CollectionName: r.collectionName,
		Vector:         vector,
		Limit:          filters.Limit,
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	results := make([]VectorSearchResult, 0, len(response.Result))
	for _, point := range response.Result {
		userID, err := r.payloadToUser(point.Payload)
		if err != nil {
			fmt.Printf("Warning: failed to convert payload to userID: %v\n", err)
			continue
		}
		results = append(results, VectorSearchResult{
			Score:  point.Score,
			UserID: userID,
		})
	}

	return results, nil
}

func (r *repository) userIDToPointID(userID string) (uint64, error) {
	var hash uint64
	for _, char := range userID {
		hash = hash*31 + uint64(char)
	}
	return hash, nil
}

func (r *repository) userToPayload(user *model.User) (map[string]*qdrant.Value, error) {
	payload := map[string]*qdrant.Value{
		"user_id": {Kind: &qdrant.Value_StringValue{StringValue: user.ID.Hex()}},
	}
	return payload, nil
}

func (r *repository) payloadToUser(payload map[string]*qdrant.Value) (string, error) {
	if v, ok := payload["user_id"]; ok {
		return v.GetStringValue(), nil
	}
	return "", fmt.Errorf("user id not found in payload")
}
