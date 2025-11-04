package service

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"io"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/adapter/qdrant"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/internal/utils/jwtUtils"
	"semki/internal/utils/mongoUtils"
	"semki/pkg/lib"
	"strconv"
	"strings"
	"sync"
	"time"
)

type searchService struct {
	qdrantService IQdrantService
	llm           ILLMService
	orgRepo       mongo.IOrganizationRepository
	chatRepo      mongo.IChatRepository
	userRepo      mongo.IUserRepository
	logger        *zap.Logger
}

// NewSearchService creates a new search service
func NewSearchService(
	vectorDB IQdrantService,
	llm ILLMService,
	orgRepo mongo.IOrganizationRepository,
	chatRepo mongo.IChatRepository,
	userRepo mongo.IUserRepository,
	logger *zap.Logger,
) routes.ISearchService {
	return &searchService{vectorDB, llm, orgRepo, chatRepo, userRepo, logger}
}

// Search godoc
//
//	@Summary		Semantic user search
//	@Description	Performs a semantic search for users using text embeddings and optional filters.
//						Results are streamed one by one with optional AI-generated descriptions.
//	@Tags			chats
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			chatId		query		string									true	"Chat ID"
//	@Param			q			query		string									false	"Search query text for semantic similarity"
//	@Param			teams		query		[]string								false	"Filter users by team names (can be multiple)"
//	@Param			levels		query		[]string								false	"Filter users by experience levels (can be multiple)"
//	@Param			locations	query		[]string								false	"Filter users by locations (can be multiple)"
//	@Param			limit		query		int										false	"Maximum number of users to return (default 5, max 20)"
//	@Success		200			{object}	dto.SearchResultWithUserAndDescription	"Streamed search results with semantic descriptions"
//	@Failure		400			{object}	map[string]string						"Invalid query parameters"
//	@Failure		401			{object}	dto.UnauthorizedResponse				"Unauthorized"
//	@Failure		500			{object}	map[string]string						"Internal server error during search or embedding"
//	@Router			/api/v1/search [get]
func (s *searchService) Search(c *gin.Context) {
	var req dto.SearchRequest
	if err := parseSearchRequest(c, &req); err != nil {
		s.logger.Error("Failed to parse search request: " + err.Error())
		lib.ResponseBadRequest(c, err, "Invalid search parameters")
		return
	}

	if req.Limit == 0 {
		req.Limit = 5
	} else if req.Limit > 20 {
		req.Limit = 20
	}

	userClaimsRaw, ok := c.Get(jwtUtils.IdentityKey)
	if !ok || userClaimsRaw == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid Claims"})
		return
	}
	claims := userClaimsRaw.(*jwtUtils.UserClaims)
	userID := claims.ID

	ctx := c.Request.Context()

	chatObjID, err := mongoUtils.StringToObjectID(req.ChatId)
	if err != nil {
		lib.ResponseBadRequest(c, err, "invalid chat id")
		return
	}

	chat, err := s.chatRepo.GetChatByID(ctx, chatObjID, userID)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "failed to fetch chat")
		return
	} else if chat == nil {
		lib.ResponseBadRequest(c, err, "chat does not exist")
		return
	}

	filters := qdrant.SearchFilters{
		Query:     req.Query,
		Teams:     req.Teams,
		Levels:    req.Levels,
		Locations: req.Locations,
		Limit:     req.Limit,
	}

	vectorSearchResults, err := s.qdrantService.SearchUsers(ctx, filters)
	if err != nil {
		s.logger.Error("Search failed: " + err.Error())
		lib.ResponseInternalServerError(c, err, "Search failed")
		return
	}

	userIDs := make([]primitive.ObjectID, 0, len(vectorSearchResults))
	for _, res := range vectorSearchResults {
		oid, err := mongoUtils.StringToObjectID(res.UserID)
		if err != nil {
			s.logger.Warn("Failed to convert userID to ObjectID: " + err.Error())
			continue
		}
		if oid == userID {
			continue
		}
		userIDs = append(userIDs, oid)
	}

	users, err := s.userRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		s.logger.Error("Failed to get users by IDs: " + err.Error())
		lib.ResponseInternalServerError(c, err, "Search failed")
		return
	}

	userMap := make(map[string]*model.User, len(users))
	for i := range users {
		userMap[users[i].ID.Hex()] = users[i]
	}

	results := make([]dto.SearchResultWithUser, 0, len(vectorSearchResults))
	for _, res := range vectorSearchResults {
		if user, ok := userMap[res.UserID]; ok {
			results = append(results, dto.SearchResultWithUser{
				Score: res.Score,
				User:  user,
			})
		}
	}

	organization, err := s.orgRepo.GetOrganizationByID(ctx, claims.OrganizationID)
	if err != nil {
		s.logger.Error("Failed to get organization: " + err.Error())
		lib.ResponseInternalServerError(c, err, "Failed to get organization")
		return
	}

	c.Stream(func(w io.Writer) bool {
		resultsChan := make(chan dto.SearchResultWithUserAndDescription)
		clientGone := c.Writer.CloseNotify()

		go func() {
			defer close(resultsChan)
			var wg sync.WaitGroup
			for _, res := range results {
				wg.Add(1)
				go func(res dto.SearchResultWithUser) {
					defer wg.Done()
					timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
					defer cancel()
					desc, err := s.llm.DescribeUser(timeoutCtx, req.Query, *organization, *res.User)
					if err != nil {
						s.logger.Warn("DescribeUser failed: " + err.Error())
						desc = "Failed to generate reasoning"
					}

					resultsChan <- dto.SearchResultWithUserAndDescription{
						SearchResultWithUser: &res,
						Description:          desc,
					}
				}(res)
			}
			wg.Wait()
		}()

		for res := range resultsChan {
			select {
			case <-clientGone:
				s.logger.Warn("Client disconnected during stream")
				return false
			default:
				c.SSEvent("result", res)
				go func(result dto.SearchResultWithUserAndDescription) {
					message := model.Message{
						Role: "assistant",
						Content: bson.M{
							"score":       result.Score,
							"user":        result.User,
							"description": result.Description,
						},
						Timestamp: time.Now(),
					}
					s.logger.Info(fmt.Sprintf("[%f] Found user: %s", result.Score, result.User.Name))
					bgCtx := context.Background()
					if err := s.chatRepo.AddChatMessages(bgCtx, chatObjID, []model.Message{message}); err != nil {
						s.logger.Error("Failed to save chat message: " + err.Error() + ". Info chatId: " + chatObjID.Hex())
					}
				}(res)
			}
		}
		return false
	})
}

// parseSearchRequest parses search parameters from query string
// Formats: ?teams=team1,team2 или ?teams[]=team1&teams[]=team2
func parseSearchRequest(ctx *gin.Context, req *dto.SearchRequest) error {
	req.Query = ctx.Query("q")
	req.ChatId = ctx.Query("chatId")

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
