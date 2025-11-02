package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"semki/internal/model"
)

// region CRUD

type DeleteOrganizationResponse struct {
	Message string `json:"message"`
}

type UpdateOrganizationResponse struct {
	Message string `json:"message"`
}

type CreateOrganizationResponse struct {
	Message string `json:"message"`
}

type GetOrganizationResponse model.Organization

type GetOrganizationUsersResponse struct {
	Users      []*model.User `json:"users"`
	TotalCount int64         `json:"totalCount"`
}

type CreateOrganizationRequest struct {
	Title string `json:"title" example:"Staffbase"`
}

func NewOrganizationFromRequest(req CreateOrganizationRequest) model.Organization {
	return model.Organization{
		ID:    primitive.NewObjectID(),
		Plan:  model.OrganizationPlans.FREE,
		Title: req.Title,
		Semantic: model.OrganizationSemantic{
			Levels:    []model.Level{},
			Teams:     []model.Team{},
			Locations: []model.Location{},
		},
		Status: model.OrganizationStatuses.ACTIVE,
	}
}

// endregion

// region Team DTOs

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type TeamResponse struct {
	Message string `json:"message"`
}

// endregion

// region Level DTOs

type CreateLevelRequest struct {
	Name        string `json:"name" binding:"required" example:"Senior"`
	Description string `json:"description" example:"Best of the best"`
}

type UpdateLevelRequest struct {
	Name        *string `json:"name,omitempty" example:"Lvl 4"`
	Description *string `json:"description,omitempty" example:"Living computer"`
}

type LevelResponse struct {
	Message string `json:"message"`
}

// endregion

// region Location DTOs

type CreateLocationRequest struct {
	Name string `json:"name" binding:"required" example:"Chemnitz"`
}

type UpdateLocationRequest struct {
	Name *string `json:"name,omitempty" example:"Berlin"`
}

type LocationResponse struct {
	Message string `json:"message"`
}

// endregion

// region PATCH Organization DTO

type PatchOrganizationRequest struct {
	Title string `json:"title,omitempty" example:"StaffAlienBase"`
}

type PatchOrganizationResponse struct {
	Message string `json:"message"`
}

// endregion
