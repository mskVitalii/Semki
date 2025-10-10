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
