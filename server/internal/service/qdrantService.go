package service

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/adapter/qdrant"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/model"
	"semki/internal/utils/jwtUtils"
	"semki/pkg/lib"
)

type IQdrantService interface {
	IndexUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	SearchUsers(ctx context.Context, filters qdrant.SearchFilters) ([]qdrant.VectorSearchResult, error)
	ReIndexReq(c *gin.Context)
	ReIndexFunc(ctx context.Context, organizationID primitive.ObjectID) (int, error)
}

type qdrantService struct {
	repo     qdrant.IQdrantRepository
	userRepo mongo.IUserRepository
	embedder IEmbedderService
}

func NewQdrantService(repo qdrant.IQdrantRepository, userRepo mongo.IUserRepository, embedder IEmbedderService) IQdrantService {
	return &qdrantService{repo, userRepo, embedder}
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

// ReIndexReq godoc
//
//	@Summary		Re-index all users
//	@Description	Retrieves all users from the database and reindexes them in Qdrant with fresh embeddings.
//	@Tags			qdrant
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Number of reindexed users"
//	@Failure		500	{object}	map[string]string		"Failed to fetch users"
//	@Router			/api/v1/reindex [post]
//	@Security		BearerAuth
func (s *qdrantService) ReIndexReq(c *gin.Context) {
	ctx := c.Request.Context()
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	claims, ok := userClaims.(*jwtUtils.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "No claims"})
		return
	}
	organizationID := claims.OrganizationID

	totalIndexed, err := s.ReIndexFunc(ctx, organizationID)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "failed to fetch users")
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("reindexed %d users", totalIndexed)})
}

func (s *qdrantService) ReIndexFunc(ctx context.Context, organizationID primitive.ObjectID) (int, error) {
	limit := 100
	page := 1
	totalIndexed := 0

	for {
		users, total, err := s.userRepo.GetUsersByOrganization(ctx, organizationID, "", page, limit)
		if err != nil {
			return totalIndexed, err
		}

		if len(users) == 0 {
			break
		}

		for _, user := range users {
			_ = s.IndexUser(ctx, user)
			totalIndexed++
		}

		if int64(page*limit) >= total {
			break
		}
		page++
	}
	return totalIndexed, nil
}
