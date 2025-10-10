package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"semki/internal/model"
	"semki/pkg/lib"
)

type DeleteUserResponse struct {
	Message string `json:"message"`
}

type UpdateUserResponse struct {
	Message string `json:"message"`
}

type CreateUserResponse struct {
	Message string `json:"message"`
}

type GetUserResponse model.User

type CreateUserRequest struct {
	Email    string `json:"email" example:"msk.vitaly@gmail.com"`
	Password string `json:"password" example:"defaultPassword"`
}

type CreateUserByGoogleProvider struct {
	Email string `json:"email"`
}

// TODO: check

func NewUserFromRequest(req CreateUserRequest) model.User {
	return model.User{
		Id:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  lib.HashPassword(req.Password),
		Providers: []model.UserProvider{model.UserProviders.Email},
		Status:    model.UserStatuses.ACTIVE,
	}
}

func NewUserFromGoogleProvider(user CreateUserByGoogleProvider) *model.User {
	return &model.User{
		Id:        primitive.NewObjectID(),
		Email:     user.Email,
		Password:  "",
		Providers: []model.UserProvider{model.UserProviders.Google},
		Status:    model.UserStatuses.ACTIVE,
	}
}
