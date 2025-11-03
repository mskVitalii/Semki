package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"semki/internal/model"
	"semki/pkg/lib"
)

type DeleteUserResponse struct {
	Message string `json:"message"`
}

type RestoreUserResponse struct {
	Message string `json:"message"`
}

type InviteUserResponse struct {
	Message string `json:"message"`
	UserId  string `json:"user_id"`
}

type UpdateUserResponse struct {
	Message string `json:"message"`
}

type CreateUserResponse struct {
	Message string     `json:"message"`
	User    model.User `json:"user"`
}

type RegisterUserResponse struct {
	Message string `json:"message"`
	Tokens  struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token,omitempty"`
		ExpiresAt    int64  `json:"expires_at"`
		CreatedAt    int64  `json:"created_at"`
	} `json:"tokens"`
}

type GetUserResponse model.User

type CreateUserRequest struct {
	Email    string `json:"email" example:"msk.vitaly@gmail.com"`
	Password string `json:"password" example:"defaultPassword"`
}

type RegisterUserRequest struct {
	Name string `json:"name" example:"Vitalii"`
	CreateUserRequest
}

type CreateUserByGoogleProvider struct {
	Email string `json:"email"`
}

// InviteUserRequest represents the request body for inviting a user
type InviteUserRequest struct {
	Email            string                 `json:"email" binding:"required,email"`
	Name             string                 `json:"name" binding:"required"`
	OrganizationRole model.OrganizationRole `json:"organizationRole" binding:"required"`
	Semantic         *UserSemanticRequest   `json:"semantic,omitempty"`
	Contact          *UserContactRequest    `json:"contact,omitempty"`
}

type UserSemanticRequest struct {
	Description string `json:"description"`
	Team        string `json:"team"`
	Level       string `json:"level"`
	Location    string `json:"location"`
}

type UserContactRequest struct {
	Slack     string `json:"slack"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
	Telegram  string `json:"telegram"`
	WhatsApp  string `json:"whatsapp"`
}

// UserToInvite converts InviteUserRequest DTO to User model
func (req *InviteUserRequest) UserToInvite(organizationId primitive.ObjectID) (*model.User, error) {
	user := &model.User{
		ID:               primitive.NewObjectID(),
		Email:            req.Email,
		Password:         "", // No password for invited users
		Name:             req.Name,
		Providers:        []model.UserProvider{model.UserProviders.Email},
		Verified:         false,
		Status:           model.UserStatuses.INVITED,
		OrganizationID:   organizationId,
		OrganizationRole: req.OrganizationRole,
	}

	// Handle optional semantic data
	if req.Semantic != nil {
		semantic := model.UserSemantic{
			Description: req.Semantic.Description,
			Location:    req.Semantic.Location,
		}

		// Convert team ObjectID if provided
		if req.Semantic.Team != "" {
			teamID, err := primitive.ObjectIDFromHex(req.Semantic.Team)
			if err != nil {
				return nil, err
			}
			semantic.Team = teamID
		}

		// Convert level ObjectID if provided
		if req.Semantic.Level != "" {
			levelID, err := primitive.ObjectIDFromHex(req.Semantic.Level)
			if err != nil {
				return nil, err
			}
			semantic.Level = levelID
		}

		user.Semantic = semantic
	}

	// Handle optional contact data
	if req.Contact != nil {
		user.Contact = model.UserContact{
			Slack:     req.Contact.Slack,
			Telephone: req.Contact.Telephone,
			Email:     req.Contact.Email,
			Telegram:  req.Contact.Telegram,
			WhatsApp:  req.Contact.WhatsApp,
		}
	}

	return user, nil
}

func NewUserFromRequest(req CreateUserRequest) *model.User {
	return &model.User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  lib.HashPassword(req.Password),
		Providers: []model.UserProvider{model.UserProviders.Email},
		Status:    model.UserStatuses.ACTIVE,
	}
}

func NewUserFromGoogleProvider(user CreateUserByGoogleProvider) *model.User {
	return &model.User{
		ID:        primitive.NewObjectID(),
		Email:     user.Email,
		Password:  "",
		Providers: []model.UserProvider{model.UserProviders.Google},
		Status:    model.UserStatuses.ACTIVE,
	}
}

type SetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type SuccessResponse struct {
	Message string `json:"message" binding:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ConfirmResetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type PatchUserRequest struct {
	Email    *string `json:"email,omitempty" bson:"email,omitempty"`
	Name     *string `json:"name,omitempty" bson:"name,omitempty"`
	Semantic *struct {
		Description *string             `json:"description,omitempty" bson:"description,omitempty"`
		Team        *primitive.ObjectID `json:"team,omitempty" bson:"team,omitempty"`
		Level       *primitive.ObjectID `json:"level,omitempty" bson:"level,omitempty"`
		Location    *string             `json:"location,omitempty" bson:"location,omitempty"`
	} `json:"semantic,omitempty" bson:"semantic,omitempty"`
	Contact *struct {
		Slack     *string `json:"slack,omitempty" bson:"slack,omitempty"`
		Telephone *string `json:"telephone,omitempty" bson:"telephone,omitempty"`
		Email     *string `json:"email,omitempty" bson:"email,omitempty"`
		Telegram  *string `json:"telegram,omitempty" bson:"telegram,omitempty"`
		WhatsApp  *string `json:"whatsapp,omitempty" bson:"whatsapp,omitempty"`
	} `json:"contact,omitempty" bson:"contact,omitempty"`
}
