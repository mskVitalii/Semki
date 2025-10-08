package model

import "go.mongodb.org/mongo-driver/bson/primitive"

//region OrganizationRole

type OrganizationRole string

var OrganizationRoles = struct {
	OWNER OrganizationRole
	ADMIN OrganizationRole
	USER  OrganizationRole
}{
	OWNER: "OWNER",
	ADMIN: "ADMIN",
	USER:  "USER",
}

//endregion

//region OrganizationPlan

type OrganizationPlanType string

var OrganizationPlans = struct {
	FREE     OrganizationPlanType
	BUSINESS OrganizationPlanType
	INVITED  OrganizationPlanType
}{
	FREE:     "FREE",
	BUSINESS: "BUSINESS",
}

//endregion

type Organization struct {
	ID       primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title    string               `bson:"title" json:"title"`
	Semantic OrganizationSemantic `bson:"semantic" json:"semantic"`
	Plan     OrganizationPlanType `bson:"plan" json:"plan"`
}

type OrganizationSemantic struct {
	LevelsIDs    []primitive.ObjectID `bson:"levelsIds" json:"levelsIds"`
	TeamsIDs     []primitive.ObjectID `bson:"teamsIds" json:"teamsIds"`
	LocationsIDs []primitive.ObjectID `bson:"locationsIds" json:"locationsIds"`
}

type Level struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganizationID primitive.ObjectID `bson:"organizationId" json:"organizationId"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description" json:"description"`
}

type Team struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganizationID primitive.ObjectID `bson:"organizationId" json:"organizationId"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description" json:"description"`
}
