package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"semki/internal/model"
)

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

type CreateOrganizationRequest struct {
	Title string `json:"title" example:"Staffbase"`
}

// TODO: check

func NewOrganizationFromRequest(req CreateOrganizationRequest) model.Organization {
	return model.Organization{
		ID:    primitive.NewObjectID(),
		Plan:  model.OrganizationPlans.FREE,
		Title: req.Title,
		Semantic: model.OrganizationSemantic{
			Levels:    []model.Level{},
			Teams:     []model.Team{},
			Locations: []primitive.ObjectID{},
		},
		Status: model.OrganizationStatuses.ACTIVE,
	}
}
