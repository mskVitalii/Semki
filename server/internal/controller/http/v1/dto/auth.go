package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"semki/internal/model"
	"time"
)

type LoginRequest struct {
	Email        string `form:"email" json:"email" binding:"required"`
	Password     string `form:"password" json:"password" binding:"required"`
	Organization string `form:"organization" json:"organization" binding:"required"`
}

type SuccessLoginResponse struct {
	Message string    `json:"message"`
	Expire  time.Time `json:"expire"`
}

type UnauthorizedResponse struct {
	Message string `json:"message"`
}

type SuccessRefreshTokenResponse struct {
	Message string    `json:"message"`
	Expire  time.Time `json:"expire"`
}

type GetUserClaimsResponse struct {
	Id               primitive.ObjectID     `json:"id"`
	OrganizationId   primitive.ObjectID     `json:"organizationId"`
	OrganizationRole model.OrganizationRole `json:"organizationRole"`
}

type SuccessLogoutResponse struct {
	Message string `json:"message"`
}

type BadRequestResponse struct {
	Error string `json:"error"`
}
