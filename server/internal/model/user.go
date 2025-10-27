package model

import "go.mongodb.org/mongo-driver/bson/primitive"

//region User Providers

type UserProvider string

var UserProviders = struct {
	Email  UserProvider
	Google UserProvider
}{
	Email:  "email",
	Google: "google",
}

func ProviderInUserProviders(provider UserProvider, providers []UserProvider) bool {
	for _, b := range providers {
		if b == provider {
			return true
		}
	}
	return false
}

//endregion

//region UserStatus

type UserStatus string

var UserStatuses = struct {
	ACTIVE  UserStatus
	DELETED UserStatus
	INVITED UserStatus
}{
	ACTIVE:  "ACTIVE",
	DELETED: "DELETED",
	INVITED: "INVITED",
}

//endregion

type UserSemantic = struct {
	Description string             `json:"description" bson:"description"`
	Team        primitive.ObjectID `json:"team" bson:"team"`
	Level       primitive.ObjectID `json:"level" bson:"level"`
	Location    string             `json:"location" bson:"location"`
}

type UserContact struct {
	Slack     string `json:"slack" bson:"slack"`
	Telephone string `json:"telephone" bson:"telephone"`
	Email     string `json:"email" bson:"email"`
	Telegram  string `json:"telegram" bson:"telegram"`
	WhatsApp  string `json:"whatsapp" bson:"whatsapp"`
}

type User struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	Email            string             `json:"email" bson:"email"`
	Password         string             `json:"password" bson:"password"`
	Name             string             `json:"name" bson:"name"`
	Providers        []UserProvider     `json:"providers" bson:"providers"`
	Verified         bool               `json:"verified" bson:"verified"`
	Status           UserStatus         `json:"status" bson:"status"`
	Semantic         UserSemantic       `json:"semantic" bson:"semantic"`
	Contact          UserContact        `json:"contact" bson:"contact"`
	AvatarID         primitive.ObjectID `json:"avatarId" bson:"avatarId"`
	OrganizationID   primitive.ObjectID `json:"organizationId" bson:"organizationId"`
	OrganizationRole OrganizationRole   `json:"organizationRole" bson:"organizationRole"`
}
