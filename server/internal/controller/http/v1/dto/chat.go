package dto

type CreateChatRequest struct {
	Message string `json:"message" binding:"required" example:"Who are you having lasagna with today and why?"`
}

type CreateChatResponse struct {
	ID        string `json:"id"`
	Title     string `bson:"title"`
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
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
