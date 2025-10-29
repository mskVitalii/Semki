package service

import (
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/internal/utils/jwtUtils"
	"semki/internal/utils/mongoUtils"
	"semki/pkg/lib"
	"semki/pkg/telemetry"
)

// userService - dependent services
type userService struct {
	repo         mongo.IRepository
	emailService EmailService
	jwtAuth      *jwt.GinJWTMiddleware
	frontendUrl  string
}

func NewUserService(repo mongo.IRepository, emailService EmailService, jwtAuth *jwt.GinJWTMiddleware, frontendUrl string) routes.IUserService {
	return &userService{repo, emailService, jwtAuth, frontendUrl}
}

// CreateUser godoc
//
//	@Summary		Creates a new user
//	@Description	Creates a new user in the MongoDB database.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			user	body		dto.CreateUserRequest	true	"User object to create"
//	@Success		201		{object}	dto.CreateUserResponse	"Successful response"
//	@Failure		400		{object}	lib.ErrorResponse		"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse		"Internal server error"
//	@Router			/api/v1/user [post]
func (s *userService) CreateUser(c *gin.Context) {
	var userDto dto.CreateUserRequest
	if err := c.ShouldBindJSON(&userDto); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	ctx := c.Request.Context()
	if lib.IsValidEmail(userDto.Email) == false {
		lib.ResponseBadRequest(c, errors.New("Invalid email"), "Invalid email")
		return
	}
	if lib.IsValidPassword(userDto.Password) == false {
		lib.ResponseBadRequest(c, errors.New("Invalid password"), "Invalid password")
		return
	}

	userByEmail, err := s.repo.GetUserByEmail(ctx, userDto.Email)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to check user existence")
		return
	}

	// Adding provider Email & Password
	if userByEmail != nil {
		if model.ProviderInUserProviders(model.UserProviders.Email, userByEmail.Providers) {
			lib.ResponseBadRequest(c, err, "User already exists")
			return
		}
		userByEmail.Password = lib.HashPassword(userDto.Password)
		userByEmail.Providers = append(userByEmail.Providers, model.UserProviders.Email)
		if err := s.repo.UpdateUser(ctx, userByEmail.ID, *userByEmail); err != nil {
			lib.ResponseInternalServerError(c, err, "Error while adding Email Provider to existing user")
			return
		}

		c.JSON(http.StatusOK, dto.CreateUserResponse{Message: "Added Email Provider to existing User", User: *userByEmail})
		return
	}

	// Creating user
	user := dto.NewUserFromRequest(userDto)

	if err := s.repo.CreateUser(ctx, user); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create user")
		return
	}

	c.JSON(http.StatusCreated, dto.CreateUserResponse{Message: "User created", User: *user})
}

// RegisterUser godoc
//
//	@Summary		Register a new user
//	@Description	Creates a new user in the MongoDB database + create organization + tokens + sends verification email
//		(vs Create to comply with CRUD / REST)
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		dto.RegisterUserRequest		true	"User object to create"
//	@Success		201		{object}	dto.RegisterUserResponse	"Successful response"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/register [post]
func (s *userService) RegisterUser(c *gin.Context) {
	var userDto dto.RegisterUserRequest
	if err := c.ShouldBindJSON(&userDto); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	ctx := c.Request.Context()
	if lib.IsValidEmail(userDto.Email) == false {
		lib.ResponseBadRequest(c, errors.New("Invalid email"), "Invalid email")
		return
	}
	if lib.IsValidPassword(userDto.Password) == false {
		lib.ResponseBadRequest(c, errors.New("Invalid password"), "Invalid password")
		return
	}

	userByEmail, err := s.repo.GetUserByEmail(ctx, userDto.Email)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to check user existence")
		return
	}

	// NEW PROVIDER / EXISTING ACCOUNT
	if userByEmail != nil {
		if model.ProviderInUserProviders(model.UserProviders.Email, userByEmail.Providers) {
			lib.ResponseBadRequest(c, err, "User already exists")
			return
		}
		userByEmail.Password = lib.HashPassword(userDto.Password)
		userByEmail.Providers = append(userByEmail.Providers, model.UserProviders.Email)
		if err := s.repo.UpdateUser(ctx, userByEmail.ID, *userByEmail); err != nil {
			lib.ResponseInternalServerError(c, err, "Error while adding Email Provider to existing user")
			return
		}

		_, err := s.repo.GetOrganizationByID(ctx, userByEmail.OrganizationID)
		if err != nil {
			lib.ResponseInternalServerError(c, err, "Failed to get organization")
			return
		}

		claims, err := jwtUtils.UserToPayload(userByEmail)
		jwtToken, err := s.jwtAuth.TokenGenerator(claims)
		if err != nil {
			telemetry.Log.Error(err.Error())
			c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20token")
			return
		}

		c.JSON(http.StatusOK, dto.RegisterUserResponse{Message: "Added Email Provider to existing User", Tokens: *jwtToken})
		return
	}

	// Creating user with organization
	organization := dto.NewOrganizationFromRequest(dto.CreateOrganizationRequest{
		Title: fmt.Sprintf("%s's organization", userDto.Name),
	})
	err = s.repo.CreateOrganization(ctx, organization)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create organization")
		return
	}

	user := dto.NewUserFromRequest(userDto.CreateUserRequest)
	user.OrganizationID = organization.ID
	user.OrganizationRole = model.OrganizationRoles.OWNER

	if err := s.repo.CreateUser(ctx, user); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create user")
		return
	}

	// Verification email
	// TODO: verification link + handler
	err = s.emailService.SendVerificationEmail(user.Email, user.Name, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: err.Error()})
		return
	}

	// Token
	claims, err := jwtUtils.UserToPayload(user)
	jwtToken, err := s.jwtAuth.TokenGenerator(claims)
	if err != nil {
		telemetry.Log.Error(err.Error())
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20token")
		return
	}
	c.JSON(http.StatusCreated, dto.RegisterUserResponse{Message: "User created", Tokens: *jwtToken})
}

// GetUser godoc
//
//	@Summary		Retrieves a user by its ID
//	@Description	Retrieves a user from the MongoDB database by its ID.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string						true	"ID of the user to retrieve"
//	@Success		200	{object}	dto.GetUserResponse			"Successful response"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		404	{object}	lib.ErrorResponse			"User not found"
//	@Failure		500	{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/{id} [get]
func (s *userService) GetUser(c *gin.Context) {
	telemetry.Log.Info("GetUser")

	id := c.Param("id")
	paramObjectId, err := mongoUtils.StringToObjectID(id)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong user id"), "Wrong id format")
		return
	}

	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	userId := userClaims.(*jwtUtils.UserClaims).ID
	if userId != paramObjectId {
		c.JSON(http.StatusForbidden, dto.UnauthorizedResponse{Message: "Forbidden"})
		return
	}
	telemetry.Log.Info(fmt.Sprintf("GetUser -> userId%s", userId))

	ctx := c.Request.Context()
	user, err := s.repo.GetUserByID(ctx, paramObjectId)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to get user")
		return
	}
	telemetry.Log.Info(fmt.Sprintf("GetUser -> user is nil %t", user == nil))

	if user == nil {
		lib.ResponseNotFound(c, "User not found")
		return
	}
	telemetry.Log.Info(fmt.Sprintf("GetUser -> user email %s", user.Email))

	if user.Status == model.UserStatuses.DELETED {
		lib.ResponseNotFound(c, "User not found")
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
//
//	@Summary		Updates a user by its ID
//	@Description	Updates a user in the MongoDB database by its ID.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"ID of the user to update"
//	@Param			user	body		model.User					true	"User object with updated data"
//	@Success		200		{object}	dto.UpdateUserResponse		"Successful response"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/{id} [put]
func (s *userService) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	paramObjectId, err := mongoUtils.StringToObjectID(id)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong user id"), "Wrong id format")
		return
	}

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	if lib.IsValidEmail(user.Email) == false {
		lib.ResponseBadRequest(c, errors.New("Invalid email"), "Invalid email")
		return
	}

	if lib.IsValidPassword(user.Password) == false && model.ProviderInUserProviders(model.UserProviders.Email, user.Providers) {
		lib.ResponseBadRequest(c, errors.New("Invalid password"), "Invalid password")
		return
	}

	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	userId := userClaims.(*jwtUtils.UserClaims).ID

	if userId != user.ID || userId != paramObjectId {
		lib.ResponseBadRequest(c, errors.New("Wrong user id"), "User id must be the same user")
		return
	}
	if user.Status == model.UserStatuses.DELETED {
		lib.ResponseBadRequest(c, errors.New("Wrong method"), "Use DELETE /api/v1/user to change user status")
		return
	}

	ctx := c.Request.Context()
	userByID, err := s.repo.GetUserByID(ctx, paramObjectId)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("User doesn't exist"), "Use correct id")
		return
	}

	// Changing password
	if userByID.Password != user.Password {
		user.Password = lib.HashPassword(userByID.Password)
	} else if user.Password == "" {
		user.Password = userByID.Password
	}

	if err := s.repo.UpdateUser(ctx, paramObjectId, user); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to update user")
		return
	}

	c.JSON(http.StatusOK, dto.UpdateUserResponse{Message: "User updated"})
}

// DeleteUser godoc
//
//	@Summary		Deletes a user by its ID
//	@Description	Deletes a user from the MongoDB database by its ID.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string						true	"ID of the user to delete"
//	@Success		200	{object}	dto.DeleteUserResponse		"Successful response"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		500	{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/{id} [delete]
func (s *userService) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	paramObjectId, err := mongoUtils.StringToObjectID(id)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong user id"), "Wrong id format")
		return
	}

	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid Claims"})
		return
	}
	userId := userClaims.(*jwtUtils.UserClaims).ID
	if userId != paramObjectId {
		lib.ResponseBadRequest(c, errors.New("Wrong user id"), "User id must be for the same user")
		return
	}

	if err := s.repo.DeleteUser(ctx, paramObjectId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to delete user")
		return
	}

	c.JSON(http.StatusOK, dto.DeleteUserResponse{Message: "User deleted"})
}

// RestoreUser godoc
//
//	@Summary		Restores a user by its ID
//	@Description	Restores a user from the MongoDB database by its ID.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string						true	"ID of the user to restore"
//	@Success		200	{object}	dto.RestoreUserResponse		"Successful response"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		500	{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/{id}/restore [post]
func (s *userService) RestoreUser(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	paramObjectId, err := mongoUtils.StringToObjectID(id)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong user id"), "Wrong id format")
		return
	}

	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Unauthorized"})
		return
	}
	claims, ok := userClaims.(*jwtUtils.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid Claims"})
		return
	}

	if claims.OrganizationRole != model.OrganizationRoles.OWNER || claims.OrganizationRole != model.OrganizationRoles.ADMIN {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "No access"})
		return
	}

	if err := s.repo.RestoreUser(ctx, paramObjectId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to restore user")
		return
	}

	c.JSON(http.StatusOK, dto.DeleteUserResponse{Message: "User deleted"})
}

// InviteUser godoc
//
//	@Summary		Invites a user by its ID
//	@Description	Invites a user from the MongoDB database by its ID.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.InviteUserResponse		"Successful response"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		500	{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/invite [post]
func (s *userService) InviteUser(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Unauthorized"})
		return
	}
	claims, ok := userClaims.(*jwtUtils.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid Claims"})
		return
	}

	if claims.OrganizationRole != model.OrganizationRoles.OWNER && claims.OrganizationRole != model.OrganizationRoles.ADMIN {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "No access"})
		return
	}

	// Bind request body
	var req dto.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, lib.ErrorResponse{Message: err.Error()})
		return
	}

	user, err := req.UserToInvite(claims.OrganizationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.ErrorResponse{Message: "Invalid data: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := s.repo.CreateUser(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: err.Error()})
		return
	}

	organization, err := s.repo.GetOrganizationByID(ctx, claims.OrganizationId)
	if err != nil || organization == nil {
		c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: "Organization not found"})
		return
	}

	// TODO: invitation link + handler
	err = s.emailService.SendInvitationEmail(user.Email, user.Name, organization.Title, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: err.Error()})
	}

	c.JSON(http.StatusOK, dto.InviteUserResponse{
		Message: "User invited successfully",
		UserId:  user.ID.Hex(),
	})
}
