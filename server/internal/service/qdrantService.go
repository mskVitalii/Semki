package service

import (
	"context"
	"fmt"
	"semki/internal/adapter/qdrant"
	"semki/internal/model"
)

type IQdrantService interface {
	IndexUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	SearchUsers(ctx context.Context, filters qdrant.SearchFilters) ([]qdrant.VectorSearchResult, error)
}

type qdrantService struct {
	repo     qdrant.IQdrantRepository
	embedder IEmbedderService
}

func NewQdrantService(repo qdrant.IQdrantRepository, embedder IEmbedderService) IQdrantService {
	return &qdrantService{repo: repo, embedder: embedder}
}

func (s *qdrantService) IndexUser(ctx context.Context, user *model.User) error {
	vector, err := s.embedder.Embed(user.Semantic.Description)
	if err != nil {
		return fmt.Errorf("embedding failed: %w", err)
	}
	return s.repo.IndexUserWithVector(ctx, user, vector)
}

func (s *qdrantService) UpdateUser(ctx context.Context, user *model.User) error {
	vector, err := s.embedder.Embed(user.Semantic.Description)
	if err != nil {
		return fmt.Errorf("embedding failed: %w", err)
	}
	return s.repo.UpdateUserWithVector(ctx, user, vector)
}

func (s *qdrantService) SearchUsers(ctx context.Context, filters qdrant.SearchFilters) ([]qdrant.VectorSearchResult, error) {
	vector, err := s.embedder.Embed(filters.Query)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}
	return s.repo.SearchUserByVector(ctx, vector, filters)
}

func (s *qdrantService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}
