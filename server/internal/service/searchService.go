package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"io"
	"semki/internal/adapter/mongo"
	"semki/internal/adapter/qdrant"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/internal/utils/mongoUtils"
	"semki/pkg/lib"
	"strconv"
	"strings"
	"sync"
)

type searchService struct {
	embedder   IEmbedderService
	qdrantRepo qdrant.IQdrantRepository
	mongoRepo  mongo.IMongoRepository
	logger     *zap.Logger
}

// NewSearchService creates a new search service
func NewSearchService(
	embedder IEmbedderService,
	qdrantRepo qdrant.IQdrantRepository,
	mongoRepo mongo.IMongoRepository,
	logger *zap.Logger,
) routes.ISearchService {
	return &searchService{
		embedder:   embedder,
		qdrantRepo: qdrantRepo,
		mongoRepo:  mongoRepo,
		logger:     logger,
	}
}

// SearchRequest represents the search query parameters
type SearchRequest struct {
	Query     string   `form:"q" json:"q"`                    // Текстовый запрос
	Teams     []string `form:"teams" json:"teams"`            // Фильтр по командам
	Levels    []string `form:"levels" json:"levels"`          // Фильтр по уровням
	Locations []string `form:"locations" json:"locations"`    // Фильтр по локациям
	Limit     uint64   `form:"limit,default=10" json:"limit"` // Лимит результатов
}

type SearchResultWithUser struct {
	Score float32     `json:"score"`
	User  *model.User `json:"user"`
}

type SearchResultWithUserAndDescription struct {
	*SearchResultWithUser
	Description string `json:"description,omitempty"`
}

// Search godoc
//
//	@Summary		Semantic user search
//	@Description	Performs a semantic search for users using text embeddings and optional filters.
//					Results are streamed one by one with optional AI-generated descriptions.
//	@Tags			search
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			q			query		string								false	"Search query text for semantic similarity"
//	@Param			teams		query		[]string							false	"Filter users by team names (can be multiple)"
//	@Param			levels		query		[]string							false	"Filter users by experience levels (can be multiple)"
//	@Param			locations	query		[]string							false	"Filter users by locations (can be multiple)"
//	@Param			limit		query		int									false	"Maximum number of users to return (default 5, max 20)"
//	@Success		200			{object}	SearchResultWithUserAndDescription	"Streamed search results with semantic descriptions"
//	@Failure		400			{object}	map[string]string					"Invalid query parameters"
//	@Failure		500			{object}	map[string]string					"Internal server error during search or embedding"
//	@Router			/api/v1/search [get]
func (s *searchService) Search(ctx *gin.Context) {
	var req SearchRequest
	if err := parseSearchRequest(ctx, &req); err != nil {
		s.logger.Error("Failed to parse search request: " + err.Error())
		lib.ResponseBadRequest(ctx, err, "Invalid search parameters")
		return
	}
	if req.Limit > 20 {
		req.Limit = 20
	} else if req.Limit == 0 {
		req.Limit = 5
	}

	vector, err := s.embedder.Embed(req.Query)
	if err != nil {
		desc := fmt.Sprintf("Failed to generate embedding: %v", err)
		s.logger.Error(desc)
		lib.ResponseInternalServerError(ctx, err, desc)
		return
	}

	filters := qdrant.SearchFilters{
		Query:     req.Query,
		Teams:     req.Teams,
		Levels:    req.Levels,
		Locations: req.Locations,
		Limit:     req.Limit,
	}

	vectorSearchResults, err := s.qdrantRepo.SearchUserByVector(ctx.Request.Context(), vector, filters)
	if err != nil {
		s.logger.Error("Search failed: " + err.Error())
		lib.ResponseInternalServerError(ctx, err, "Search failed")
		return
	}

	userIDs := make([]primitive.ObjectID, 0, len(vectorSearchResults))
	for _, res := range vectorSearchResults {
		oid, err := mongoUtils.StringToObjectID(res.UserID)
		if err != nil {
			s.logger.Warn("Failed to convert userID to ObjectID: " + err.Error())
			continue
		}
		userIDs = append(userIDs, oid)
	}

	users, err := s.mongoRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		s.logger.Error("Failed to get users by IDs: " + err.Error())
		lib.ResponseInternalServerError(ctx, err, "Search failed")
		return
	}

	// TODO: use LLMService to get the description
	// TODO: stream the SearchResultWithUser one-by-one in goroutines
	results := make([]SearchResultWithUser, 0, len(users))
	for _, res := range vectorSearchResults {
		oid, err := mongoUtils.StringToObjectID(res.UserID)
		if err != nil {
			continue
		}
		for _, u := range users {
			if u.Id == oid {
				results = append(results, SearchResultWithUser{
					Score: res.Score,
					User:  u,
				})
				break
			}
		}
	}

	ctx.Stream(func(w io.Writer) bool {
		resultsChan := make(chan SearchResultWithUserAndDescription)
		go func() {
			defer close(resultsChan)
			var wg sync.WaitGroup
			for _, res := range results {
				wg.Add(1)
				go func(res SearchResultWithUser) {
					defer wg.Done()

					desc := "TODO s.llmService.DescribeUser(ctx, user)"
					//desc, _ := s.llmService.DescribeUser(ctx, user)
					resultsChan <- SearchResultWithUserAndDescription{
						SearchResultWithUser: &SearchResultWithUser{
							Score: res.Score,
							User:  res.User,
						},
						Description: desc,
					}
				}(res)
			}
			wg.Wait()
		}()

		for res := range resultsChan {
			ctx.SSEvent("result", res)
		}

		return false
	})
}

// parseSearchRequest parses search parameters from query string
// Formats: ?teams=team1,team2 или ?teams[]=team1&teams[]=team2
func parseSearchRequest(ctx *gin.Context, req *SearchRequest) error {
	req.Query = ctx.Query("q")

	// Teams
	if teams := ctx.Query("teams"); teams != "" {
		req.Teams = strings.Split(teams, ",")
	} else {
		req.Teams = ctx.QueryArray("teams[]")
	}

	// Levels
	if levels := ctx.Query("levels"); levels != "" {
		req.Levels = strings.Split(levels, ",")
	} else {
		req.Levels = ctx.QueryArray("levels[]")
	}

	// Locations
	if locations := ctx.Query("locations"); locations != "" {
		req.Locations = strings.Split(locations, ",")
	} else {
		req.Locations = ctx.QueryArray("locations[]")
	}

	// Limit
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limit, err := strconv.ParseUint(limitStr, 10, 64)
		if err != nil {
			return err
		}
		req.Limit = limit
	} else {
		req.Limit = 5
	}

	req.Teams = filterEmpty(req.Teams)
	req.Levels = filterEmpty(req.Levels)
	req.Locations = filterEmpty(req.Locations)

	return nil
}

// filterEmpty removes empty strings from slice
func filterEmpty(items []string) []string {
	var result []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
