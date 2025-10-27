package service

import (
	"context"
	"errors"
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/pkg/lib"
)

// authService - dependent services
type authService struct {
	repo mongo.IRepository
}

func NewAuthService(repo mongo.IRepository) routes.IAuthService {
	return &authService{repo}
}

// RefreshTokenHandler godoc
//
//	@Summary		Refreshes the authentication token
//	@Description	Generates a new authentication token using the refresh token provided.
//	@Tags			auth
//	@Accept			json
//	@Param			refresh_token	body	dto.RefreshTokenRequest	true	"Refresh token"
//	@Produce		json
//	@Security		BearerAuth
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Router			/api/v1/refresh_token [post]
func (s *authService) RefreshTokenHandler(_ *gin.Context) {
	// This method for swagger. Token refresh happens inside jwtUtils.go
}

// LogoutHandler godoc
//
//	@Summary		Logs out a user and invalidates the JWT token
//	@Description	Invalidates the JWT token for the user by adding it to the blacklist and removes the JWT cookie.
//	@Tags			auth
//	@Param			refresh_token	body	dto.LogoutRequest	true	"Refresh token"
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.SuccessLogoutResponse	"Successful logout"
//	@Failure		400	{object}	lib.ErrorResponse			"Bad Request"
//	@Router			/api/v1/logout [post]
func (s *authService) LogoutHandler(_ *gin.Context) {
	// This method for swagger
}

// LoginHandler godoc
//
//	@Summary		Authenticates a user and returns a JWT token
//	@Description	Authenticates a user with email and password, and returns a JWT token if the credentials are valid.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest			true	"Login request body"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Router			/api/v1/login [post]
func (s *authService) LoginHandler(_ *gin.Context) {
	// This method for swagger
}

func (s *authService) Authenticate(request dto.LoginRequest) (*model.User, error) {
	ctx := context.Background()
	if lib.IsValidEmail(request.Email) == false {
		return nil, errors.New("invalid email")
	}
	if lib.IsValidPassword(request.Password) == false {
		return nil, errors.New("invalid password")
	}

	user, err := s.repo.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.Status == model.UserStatuses.DELETED {
		return nil, errors.New("user is deleted")
	}

	if model.ProviderInUserProviders(model.UserProviders.Email, user.Providers) == false {
		return nil, errors.New("invalid provider")
	}

	if lib.CheckPasswordHash(request.Password, user.Password) {
		return user, nil
	} else {
		return nil, errors.New("password is not match")
	}
}

// ClaimsHandler godoc
//
//	@Summary		Retrieves user claims
//	@Description	Fetches and returns the claims of the authenticated user.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	jwtUtils.UserClaims			"Successful response with user claims"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"No claims found"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Invalid claims type"
//	@Failure		500	{object}	dto.UnauthorizedResponse	"Internal server error"
//	@Router			/api/v1/claims [get]
func (s *authService) ClaimsHandler(c *gin.Context) {
	claimsRaw := jwt.ExtractClaims(c)
	c.JSON(http.StatusOK, claimsRaw)
}
