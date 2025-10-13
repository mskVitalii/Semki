package dto

import "semki/internal/model"

type SearchUsersRequest struct {
	Teams         []string `json:"teams,omitempty"`
	Levels        []string `json:"levels,omitempty"`
	IsBarrierFree bool     `json:"isBarrierFree,omitempty,string"`
}

// TODO: return in stream

type FoundUser struct {
	User    model.User `json:"user"`
	Message string     `json:"message"`
	Match   float32    `json:"match"`
}

type SearchUsersResponse struct {
	Data []FoundUser `json:"data"`
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
