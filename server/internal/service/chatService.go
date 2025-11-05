package service

import (
	"fmt"
	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/internal/utils/jwtUtils"
	"semki/internal/utils/mongoUtils"
	"semki/pkg/lib"
	"semki/pkg/telemetry"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type chatService struct {
	chatRepo mongo.IChatRepository
	userRepo mongo.IUserRepository
}

func NewChatService(chatRepo mongo.IChatRepository, userRepo mongo.IUserRepository) routes.IChatService {
	return &chatService{chatRepo, userRepo}
}

// CreateChat
//
//	@Summary	Create new chat
//	@Tags		chats
//	@Accept		json
//	@Produce	json
//	@Param		request	body		dto.CreateChatRequest	true	"Chat creation request"
//	@Success	201		{object}	dto.CreateChatResponse
//	@Failure	400		{object}	map[string]string
//	@Failure	401		{object}	map[string]string
//	@Failure	500		{object}	map[string]string
//	@Security	BearerAuth
//	@Router		/chat [post]
func (s *chatService) CreateChat(c *gin.Context) {
	var req dto.CreateChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "invalid request body")
		return
	}

	userClaims, claimsExists := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil || claimsExists == false {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid Claims"})
		return
	}
	userID := userClaims.(*jwtUtils.UserClaims).ID

	chat := &model.Chat{
		UserID: userID,
		Title:  req.Message,
		Messages: []model.Message{
			{
				Role:      openai.ChatMessageRoleUser,
				Content:   bson.M{"title": req.Message},
				Timestamp: time.Now(),
			},
		},
	}

	if err := s.chatRepo.CreateChat(c.Request.Context(), chat); err != nil {
		lib.ResponseInternalServerError(c, err, "failed to create chat")
		return
	}

	response := dto.CreateChatResponse{
		ID:        chat.ID.Hex(),
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt.Unix(),
	}

	c.JSON(http.StatusCreated, response)
}

// GetChat
//
//	@Summary	Get chat by ID
//	@Tags		chats
//	@Produce	json
//	@Param		id	path		string	true	"Chat ID"
//	@Success	200	{object}	dto.GetChatResponse
//	@Failure	400	{object}	map[string]string
//	@Failure	401	{object}	map[string]string
//	@Failure	404	{object}	map[string]string
//	@Failure	500	{object}	map[string]string
//	@Security	BearerAuth
//	@Router		/chat/{id} [get]
func (s *chatService) GetChat(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		lib.ResponseBadRequest(c, nil, "id is required")
		return
	}

	userClaims, claimsExists := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil || claimsExists == false {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid Claims"})
		return
	}
	userID := userClaims.(*jwtUtils.UserClaims).ID

	chatObjID, err := mongoUtils.StringToObjectID(id)
	if err != nil {
		lib.ResponseBadRequest(c, err, "invalid chat id")
		return
	}

	ctx := c.Request.Context()
	chat, err := s.chatRepo.GetChatByID(ctx, chatObjID, userID)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "failed to fetch chat")
		return
	}

	if chat == nil {
		lib.ResponseNotFound(c, "chat not found")
		return
	}

	idsMap := make(map[primitive.ObjectID]struct{})
	for _, msg := range chat.Messages {
		if uid, ok := msg.Content["user"].(primitive.ObjectID); ok {
			idsMap[uid] = struct{}{}
		} else if uidStr, ok := msg.Content["user"].(string); ok {
			uidObj, err := primitive.ObjectIDFromHex(uidStr)
			if err == nil {
				idsMap[uidObj] = struct{}{}
			}
		}
	}

	ids := make([]primitive.ObjectID, 0, len(idsMap))
	for id := range idsMap {
		ids = append(ids, id)
	}

	users, err := s.userRepo.GetUsersByIDs(ctx, ids)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Cannot get chat users")
		return
	}

	userMap := make(map[string]interface{}, len(users))
	for _, u := range users {
		userMap[u.ID.Hex()] = u
	}

	messages := make([]map[string]interface{}, 0, len(chat.Messages))
	for _, msg := range chat.Messages {
		content := make(map[string]interface{})
		for k, v := range msg.Content {
			telemetry.Log.Info(fmt.Sprintf("key: %s", k))
			if k == "user" {
				telemetry.Log.Info(fmt.Sprintf("%s", v))
				if u, ok := userMap[v.(string)]; ok {
					telemetry.Log.Info(fmt.Sprintf("Found userMap by %s", v.(string)))

					content[k] = u
				} else {
					content[k] = v
				}
			} else {
				content[k] = v
			}
		}
		messages = append(messages, content)
		telemetry.Log.Info(fmt.Sprintf("new Len: %d", len(messages)))
	}

	response := dto.GetChatResponse{
		ID:        chat.ID.Hex(),
		Messages:  messages,
		CreatedAt: chat.CreatedAt.Unix(),
		UpdatedAt: chat.UpdatedAt.Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// GetUserHistory
//
//	@Summary	Get user chat history
//	@Tags		chats
//	@Produce	json
//	@Param		cursor	query		string	false	"Cursor for pagination"
//	@Param		limit	query		int		false	"Number of items per page"	default(20)
//	@Success	200		{object}	dto.GetUserHistoryResponse
//	@Failure	401		{object}	map[string]string
//	@Failure	500		{object}	map[string]string
//	@Security	BearerAuth
//	@Router		/chat/history [get]
func (s *chatService) GetUserHistory(c *gin.Context) {
	userClaims, claimsExists := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil || claimsExists == false {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid Claims"})
		return
	}
	userID := userClaims.(*jwtUtils.UserClaims).ID

	// Pagination
	cursor := c.Query("cursor")
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	chatsData, nextCursor, err := s.chatRepo.GetChatsByUserIDWithCursor(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "failed to fetch history")
		return
	}

	chats := make([]dto.ChatHistoryItem, 0, len(chatsData))
	for _, chat := range chatsData {
		chats = append(chats, dto.ChatHistoryItem{
			ID:        chat.ID.Hex(),
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt.Unix(),
			UpdatedAt: chat.UpdatedAt.Unix(),
		})
	}

	c.JSON(http.StatusOK, dto.GetUserHistoryResponse{
		Chats:      chats,
		NextCursor: nextCursor,
	})
}
