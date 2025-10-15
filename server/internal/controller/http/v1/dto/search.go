package dto

import "semki/internal/model"

// SearchRequest represents the search query parameters
type SearchRequest struct {
	Query     string   `form:"q" json:"q"`
	Teams     []string `form:"teams" json:"teams"`
	Levels    []string `form:"levels" json:"levels"`
	Locations []string `form:"locations" json:"locations"`
	Limit     uint64   `form:"limit,default=10" json:"limit"`
}

type SearchResultWithUser struct {
	Score float32     `json:"score"`
	User  *model.User `json:"user"`
}

type SearchResultWithUserAndDescription struct {
	*SearchResultWithUser
	Description string `json:"description,omitempty"`
}
