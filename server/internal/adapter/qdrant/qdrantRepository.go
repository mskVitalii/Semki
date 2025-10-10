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

const (
	UsersCollection = "users"
)

// SearchFilters represents search parameters
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
	IndexUser(ctx context.Context, user model.User) error
	UpdateUser(ctx context.Context, user model.User) error
	DeleteUser(ctx context.Context, id string) error
	SearchUserByVector(ctx context.Context, vector []float32, filter SearchFilters) ([]VectorSearchResult, error)
	InitializeCollection(ctx context.Context) error
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
	for _, collection := range collections.Collections {
		if collection.Name == r.collectionName {
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

		// Создаем индексы для полей payload
		err = r.createPayloadIndexes(ctx)
		if err != nil {
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
		//{"created_at", qdrant.FieldType_FieldTypeInteger},
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

func (r *repository) IndexUser(ctx context.Context, user model.User) error {
	// TODO: get vectors from Embedder
	vector, err := r.generateUserVector(user)
	if err != nil {
		return fmt.Errorf("failed to generate user vector: %w", err)
	}

	pointID, err := r.userIDToPointID(user.Id.Hex())
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
				Vector: &qdrant.Vector{
					Data: vector,
				},
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

func (r *repository) UpdateUser(ctx context.Context, user model.User) error {
	return r.IndexUser(ctx, user)
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
						{
							PointIdOptions: &qdrant.PointId_Num{
								Num: pointID,
							},
						},
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
	// TODO: use the filters
	response, err := r.client.Points.Search(ctx, &qdrant.SearchPoints{
		CollectionName: r.collectionName,
		Vector:         vector,
		Limit:          filters.Limit,
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	results := make([]VectorSearchResult, 0, len(response.Result))
	for _, point := range response.Result {
		userId, err := r.payloadToUser(point.Payload)
		if err != nil {
			fmt.Printf("Warning: failed to convert payload to userId: %v\n", err)
			continue
		}
		results = append(results, VectorSearchResult{
			Score:  point.Score,
			UserID: userId,
		})
	}

	return results, nil
}

func (r *repository) generateUserVector(user model.User) ([]float32, error) {
	// TODO: Get the real Vector as param
	vector := make([]float32, r.vectorSize)

	data := fmt.Sprintf("%s:%s:%s", user.Id.Hex(), user.Email, user.Name)
	hash := 0
	for _, char := range data {
		hash = (hash*31 + int(char)) % 1000000
	}

	for i := range vector {
		vector[i] = float32((hash+i)%100) / 100.0
	}

	return vector, nil
}

// userIDToPointID Hex string ID to Qdrant ID
func (r *repository) userIDToPointID(userID string) (uint64, error) {
	hash := uint64(0)
	for _, char := range userID {
		hash = hash*31 + uint64(char)
	}

	return hash, nil
}

// region Payload

func (r *repository) userToPayload(user model.User) (map[string]*qdrant.Value, error) {

	// Payload
	payload := map[string]*qdrant.Value{
		"user_id": {
			Kind: &qdrant.Value_StringValue{
				StringValue: user.Id.Hex(),
			},
		},
		//"created_at": {
		//	Kind: &qdrant.Value_IntegerValue{
		//		IntegerValue: user.CreatedAt.Unix(),
		//	},
		//},
	}

	return payload, nil
}

// payloadToUser payload from Qdrant -> user ID
func (r *repository) payloadToUser(payload map[string]*qdrant.Value) (string, error) {
	if v, ok := payload["user_id"]; ok {
		return v.GetStringValue(), nil
	}
	return "", fmt.Errorf("user id not found")
}

// endregion
