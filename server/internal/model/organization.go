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

//region UserStatus

type OrganizationStatus string

var OrganizationStatuses = struct {
	ACTIVE  OrganizationStatus
	DELETED OrganizationStatus
}{
	ACTIVE:  "ACTIVE",
	DELETED: "DELETED",
}

//endregion

//region OrganizationPlan

type OrganizationPlanType string

var OrganizationPlans = struct {
	FREE     OrganizationPlanType
	BUSINESS OrganizationPlanType
}{
	FREE:     "FREE",
	BUSINESS: "BUSINESS",
}

//endregion

type Organization struct {
	Id       primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title    string               `bson:"title" json:"title"`
	Semantic OrganizationSemantic `bson:"semantic" json:"semantic"`
	Plan     OrganizationPlanType `bson:"plan" json:"plan"`
	Status   OrganizationStatus   `bson:"status" json:"status"`
}

type OrganizationSemantic struct {
	Levels    []Level              `bson:"levels" json:"levels"`
	Teams     []Team               `bson:"teams" json:"teams"`
	Locations []primitive.ObjectID `bson:"locations" json:"locations"`
}

type Level struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
}

type Team struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
}
