package service

import (
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/internal/utils/jwtUtils"
	"semki/internal/utils/mongoUtils"
	"semki/pkg/lib"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type chatService struct {
	repo mongo.IRepository
}

func NewChatService(repo mongo.IRepository) routes.IChatService {
	return &chatService{repo}
}

// CreateChat
//
//	@Summary	Create new chat
//	@Tags		chat
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
				Role:      "user",
				Content:   bson.M{"Question": req.Message},
				Timestamp: time.Now(),
			},
		},
	}

	if err := s.repo.CreateChat(c.Request.Context(), chat); err != nil {
		lib.ResponseInternalServerError(c, err, "failed to create chat")
		return
	}

	response := dto.CreateChatResponse{
		ID:        chat.ID.Hex(),
		Message:   req.Message,
		Response:  "AI response here",
		CreatedAt: chat.CreatedAt.Unix(),
	}

	c.JSON(http.StatusCreated, response)
}

// GetChat
//
//	@Summary	Get chat by ID
//	@Tags		chat
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

	chat, err := s.repo.GetChatByID(c.Request.Context(), chatObjID, userID)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "failed to fetch chat")
		return
	}

	if chat == nil {
		lib.ResponseNotFound(c, "chat not found")
		return
	}

	messages := make([]map[string]interface{}, 0, len(chat.Messages))
	for _, msg := range chat.Messages {
		messages = append(messages, msg.Content)
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
//	@Tags		chat
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

	// Получаем параметры пагинации
	cursor := c.Query("cursor")
	limit := 20 // по умолчанию
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	chatsData, nextCursor, err := s.repo.GetChatsByUserIDWithCursor(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "failed to fetch history")
		return
	}

	chats := make([]dto.ChatHistoryItem, 0, len(chatsData))
	for _, chat := range chatsData {
		chats = append(chats, dto.ChatHistoryItem{
			ID:        chat.ID.Hex(),
			Question:  chat.Title,
			CreatedAt: chat.CreatedAt.Unix(),
			UpdatedAt: chat.UpdatedAt.Unix(),
		})
	}

	c.JSON(http.StatusOK, dto.GetUserHistoryResponse{
		Chats:      chats,
		NextCursor: nextCursor, // добавь это поле в DTO
	})
}
