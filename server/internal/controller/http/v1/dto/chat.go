package dto

type CreateChatRequest struct {
	Message string `json:"message" binding:"required"`
}

type CreateChatResponse struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Response  string `json:"response"`
	CreatedAt int64  `json:"created_at"`
}

type GetChatResponse struct {
	ID        string                   `json:"id"`
	Messages  []map[string]interface{} `json:"messages"`
	CreatedAt int64                    `json:"created_at"`
	UpdatedAt int64                    `json:"updated_at"`
}

type GetUserHistoryResponse struct {
	Chats      []ChatHistoryItem `json:"chats"`
	NextCursor string            `json:"nextCursor,omitempty"`
}

type ChatHistoryItem struct {
	ID        string `json:"id"`
	Question  string `json:"question"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
