package service

import (
	"fmt"
	ginJwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/internal/utils/config"
	"semki/internal/utils/crypto"
	"semki/internal/utils/jwtUtils"
	"semki/internal/utils/mongoUtils"
	"semki/pkg/lib"
	"semki/pkg/telemetry"
	"time"
)

// userService - dependent services
type userService struct {
	qdrantService IQdrantService
	userRepo      mongo.IUserRepository
	orgRepo       mongo.IOrganizationRepository
	emailService  *EmailService
	jwtAuth       *ginJwt.GinJWTMiddleware
	cfg           *config.Config
}

func NewUserService(qdrantService IQdrantService, userRepo mongo.IUserRepository, orgRepo mongo.IOrganizationRepository, emailService *EmailService, jwtAuth *ginJwt.GinJWTMiddleware, cfg *config.Config) routes.IUserService {
	return &userService{qdrantService, userRepo, orgRepo, emailService, jwtAuth, cfg}
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

	userByEmail, err := s.userRepo.GetUserByEmail(ctx, userDto.Email)
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
		if err := s.userRepo.UpdateUser(ctx, userByEmail.ID, *userByEmail); err != nil {
			lib.ResponseInternalServerError(c, err, "Error while adding Email Provider to existing user")
			return
		}

		c.JSON(http.StatusOK, dto.CreateUserResponse{Message: "Added Email Provider to existing User", User: *userByEmail})
		return
	}

	// Creating user
	user := dto.NewUserFromRequest(userDto)

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create user")
		return
	}

	if err := s.qdrantService.IndexUser(ctx, user); err != nil {
		telemetry.Log.Error("Failed to index user in Qdrant: " + err.Error())
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

	userByEmail, err := s.userRepo.GetUserByEmail(ctx, userDto.Email)
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
		if err := s.userRepo.UpdateUser(ctx, userByEmail.ID, *userByEmail); err != nil {
			lib.ResponseInternalServerError(c, err, "Error while adding Email Provider to existing user")
			return
		}

		_, err := s.orgRepo.GetOrganizationByID(ctx, userByEmail.OrganizationID)
		if err != nil {
			lib.ResponseInternalServerError(c, err, "Failed to get organization")
			return
		}

		jwtToken, err := s.jwtAuth.TokenGenerator(userByEmail)
		if err != nil {
			telemetry.Log.Error(err.Error())
			c.Redirect(http.StatusFound, s.cfg.FrontendUrl+"/login?error=internal%20error%20token")
			return
		}

		c.JSON(http.StatusOK, dto.RegisterUserResponse{Message: "Added Email Provider to existing User", Tokens: *jwtToken})
		return
	}

	// Creating user with organization
	organization := dto.NewOrganizationFromRequest(dto.CreateOrganizationRequest{
		Title: fmt.Sprintf("%s's organization", userDto.Name),
	})
	err = s.orgRepo.CreateOrganization(ctx, organization)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create organization")
		return
	}

	user := dto.NewUserFromRequest(userDto.CreateUserRequest)
	user.OrganizationID = organization.ID
	user.OrganizationRole = model.OrganizationRoles.OWNER

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create user")
		return
	}

	// Verification email
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, InviteClaims{
		UserID:         user.ID,
		OrganizationID: user.OrganizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(48 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	tokenString, _ := token.SignedString([]byte(s.cfg.SecretKeyJWT + "verify_user"))

	verificationLink := fmt.Sprintf("%s/api/v1/user/%s/verify/accept?token=%s", s.cfg.Protocol+"://"+s.cfg.Host+":"+s.cfg.Port, user.ID.Hex(), tokenString)

	err = s.emailService.SendVerificationEmail(user.Email, user.Name, verificationLink)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: err.Error()})
		return
	}

	if err := s.qdrantService.IndexUser(ctx, user); err != nil {
		telemetry.Log.Error("Failed to index created user in Qdrant: " + err.Error())
	}

	// Token
	jwtToken, err := s.jwtAuth.TokenGenerator(user)
	if err != nil {
		telemetry.Log.Error(err.Error())
		c.Redirect(http.StatusFound, s.cfg.FrontendUrl+"/login?error=internal%20error%20token")
		return
	}
	c.JSON(http.StatusCreated, dto.RegisterUserResponse{Message: "User created", Tokens: *jwtToken})
}

type VerifyClaims struct {
	UserID         primitive.ObjectID `json:"userId"`
	OrganizationID primitive.ObjectID `json:"organizationId"`
	jwt.RegisteredClaims
}

func (s *userService) VerifyUserEmailHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, lib.ErrorResponse{Message: "Missing token"})
		return
	}

	claims := &VerifyClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.SecretKeyJWT + "verify_user"), nil
	})
	if err != nil || !parsedToken.Valid {
		c.JSON(http.StatusUnauthorized, lib.ErrorResponse{Message: "Invalid or expired token"})
		return
	}

	ctx := c.Request.Context()
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, lib.ErrorResponse{Message: "User not found"})
		return
	}

	if user.Status == model.UserStatuses.INVITED {
		user.Status = model.UserStatuses.ACTIVE
		user.Verified = true
		err = s.userRepo.UpdateUser(ctx, user.ID, *user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: err.Error()})
			return
		}
	}

	// Token
	jwtToken, err := s.jwtAuth.TokenGenerator(user)
	if err != nil {
		telemetry.Log.Error(err.Error())
		c.Redirect(http.StatusFound, s.cfg.FrontendUrl+"/login?error=internal%20error%20token")
		return
	}

	c.Redirect(http.StatusFound, s.cfg.FrontendUrl+"/login?accessToken="+jwtToken.AccessToken+"&refreshToken="+jwtToken.RefreshToken)
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
	paramObjectID, err := mongoUtils.StringToObjectID(id)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong user id"), "Wrong id format")
		return
	}

	ctx := c.Request.Context()
	user, err := s.userRepo.GetUserByID(ctx, paramObjectID)
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
	userByID, err := s.userRepo.GetUserByID(ctx, paramObjectId)
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

	if err := s.userRepo.UpdateUser(ctx, paramObjectId, user); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to update user")
		return
	}

	if err := s.qdrantService.UpdateUser(ctx, &user); err != nil {
		telemetry.Log.Error("Failed to update user in Qdrant: " + err.Error())
	}

	c.JSON(http.StatusOK, dto.UpdateUserResponse{Message: "User updated"})
}

// PatchUser godoc
//
//	@Summary		Patches user fields by its ID
//	@Description	Partially updates a user in the MongoDB database by its ID.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"ID of the user to patch"
//	@Param			user	body		dto.PatchUserRequest		true	"Partial user data to update"
//	@Success		200		{object}	dto.UpdateUserResponse		"Successful response"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/{id} [patch]
func (s *userService) PatchUser(c *gin.Context) {
	id := c.Param("id")
	paramObjectId, err := mongoUtils.StringToObjectID(id)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong user id"), "Wrong id format")
		return
	}

	var body dto.PatchUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	claims := userClaims.(*jwtUtils.UserClaims)
	userId := claims.ID
	if userId != paramObjectId && claims.OrganizationRole == model.OrganizationRoles.USER {
		lib.ResponseBadRequest(c, errors.New("Wrong user id"), "User id must be the same user")
		return
	}

	ctx := c.Request.Context()
	existing, err := s.userRepo.GetUserByID(ctx, paramObjectId)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("User doesn't exist"), "Use correct id")
		return
	}
	if existing.Status == model.UserStatuses.DELETED {
		lib.ResponseBadRequest(c, errors.New("Wrong method"), "Use PATCH /api/v1/user/{id}/restore to change user status")
		return
	}

	update := bson.M{}
	if body.Email != nil {
		update["email"] = *body.Email
		existing.Email = *body.Email
	}
	if body.Name != nil {
		update["name"] = *body.Name
		existing.Name = *body.Name
	}
	if body.Semantic != nil && body.Semantic.Description != nil {
		desc := *body.Semantic.Description
		encrypted, err := crypto.EncryptField(desc, s.cfg.CryptoKey)
		if err != nil {
			lib.ResponseInternalServerError(c, err, "Failed to encrypt fields")
			return
		}
		update["semantic.description"] = encrypted
		existing.Semantic.Description = desc
	}

	if len(update) == 0 {
		lib.ResponseBadRequest(c, errors.New("No fields provided"), "Empty patch request")
		return
	}

	if err := s.userRepo.PatchUser(ctx, paramObjectId, update); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to patch user")
		return
	}

	if err := s.qdrantService.UpdateUser(ctx, existing); err != nil {
		telemetry.Log.Error("Failed to update user in Qdrant: " + err.Error())
	}

	c.JSON(http.StatusOK, dto.UpdateUserResponse{Message: "User patched"})
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
	claims, ok := userClaims.(*jwtUtils.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid Claims"})
		return
	}

	if paramObjectId == claims.ID {
		lib.ResponseBadRequest(c, errors.New("Cannot delete yourself"), "Cannot delete yourself")
		return
	}

	if err := s.userRepo.DeleteUser(ctx, paramObjectId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to delete user")
		return
	}

	if err := s.qdrantService.DeleteUser(ctx, id); err != nil {
		telemetry.Log.Error("Failed to delete user in Qdrant: " + err.Error())
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

	if err := s.userRepo.RestoreUser(ctx, paramObjectId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to restore user")
		return
	}

	user, err := s.userRepo.GetUserByID(ctx, paramObjectId)
	if err != nil || user == nil {
		lib.ResponseInternalServerError(c, err, "Failed to fetch restored user")
		return
	}

	if err := s.qdrantService.IndexUser(ctx, user); err != nil {
		telemetry.Log.Error("Failed to re-index user in Qdrant: " + err.Error())
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

	// Bind request body
	var req dto.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, lib.ErrorResponse{Message: err.Error()})
		return
	}

	user, err := req.UserToInvite(claims.OrganizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.ErrorResponse{Message: "Invalid data: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: err.Error()})
		return
	}

	organization, err := s.orgRepo.GetOrganizationByID(ctx, claims.OrganizationID)
	if err != nil || organization == nil {
		c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: "Organization not found"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, InviteClaims{
		UserID:         user.ID,
		OrganizationID: claims.OrganizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	tokenString, _ := token.SignedString([]byte(s.cfg.SecretKeyJWT + "invite-user"))

	invitationLink := fmt.Sprintf("%s/api/v1/user/%s/invite/accept?token=%s", s.cfg.Protocol+"://"+s.cfg.Host+":"+s.cfg.Port, user.ID.Hex(), tokenString)

	err = s.emailService.SendInvitationEmail(user.Email, user.Name, organization.Title, invitationLink)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: err.Error()})
	}

	c.JSON(http.StatusOK, dto.InviteUserResponse{
		Message: "User invited successfully",
		UserId:  user.ID.Hex(),
	})
}

type InviteClaims struct {
	UserID         primitive.ObjectID `json:"userId"`
	OrganizationID primitive.ObjectID `json:"organizationId"`
	jwt.RegisteredClaims
}

func (s *userService) InviteUserAcceptHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, lib.ErrorResponse{Message: "Missing token"})
		return
	}

	claims := &InviteClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.SecretKeyJWT + "invite-user"), nil
	})
	if err != nil || !parsedToken.Valid {
		c.JSON(http.StatusUnauthorized, lib.ErrorResponse{Message: "Invalid or expired token"})
		return
	}

	ctx := c.Request.Context()
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, lib.ErrorResponse{Message: "User not found"})
		return
	}

	if user.Status != model.UserStatuses.INVITED && user.Password != "" {
		c.JSON(http.StatusBadRequest, lib.ErrorResponse{Message: "User already activated"})
		return
	}

	if user.Status == model.UserStatuses.INVITED {
		user.Status = model.UserStatuses.ACTIVE
		user.Verified = true
		err = s.userRepo.UpdateUser(ctx, user.ID, *user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, lib.ErrorResponse{Message: err.Error()})
			return
		}
	}

	if err := s.qdrantService.IndexUser(ctx, user); err != nil {
		telemetry.Log.Error("Failed to index invited user in Qdrant: " + err.Error())
	}

	// Token
	jwtToken, err := s.jwtAuth.TokenGenerator(user)
	if err != nil {
		telemetry.Log.Error(err.Error())
		c.Redirect(http.StatusFound, s.cfg.FrontendUrl+"/onboarding?error=internal%20error%20token")
		return
	}

	c.Redirect(http.StatusFound, s.cfg.FrontendUrl+"/onboarding?accessToken="+jwtToken.AccessToken+"&refreshToken="+jwtToken.RefreshToken)
}

// SetPassword godoc
//
//	@Summary		Sets password for invited user
//	@Description	Accepts invitation token and sets user password
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			data	body		dto.SetPasswordRequest		true	"Token and new password"
//	@Success		200		{object}	dto.SuccessResponse			"Password set successfully"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/set_password [post]
func (s *userService) SetPassword(c *gin.Context) {
	var req dto.SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	ctx := c.Request.Context()

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
	userID := claims.ID

	if lib.IsValidPassword(req.Password) == false {
		lib.ResponseBadRequest(c, errors.New("Invalid password"), "Password must meet requirements")
		return
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("User not found"), "Use correct invitation link")
		return
	}

	if user.Password != "" {
		lib.ResponseBadRequest(c, errors.New("User already active"), "Password already set")
		return
	}

	user.Password = lib.HashPassword(req.Password)
	user.Status = model.UserStatuses.ACTIVE

	if err := s.userRepo.UpdateUser(ctx, userID, *user); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to update user")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "Password set successfully"})
}

type ResetPasswordClaims struct {
	UserID         primitive.ObjectID `json:"userId"`
	OrganizationID primitive.ObjectID `json:"organizationId"`
	jwt.RegisteredClaims
}

// ResetPassword godoc
//
//	@Summary		Request password reset
//	@Description	Sends a password reset email with a secure link to the user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			data	body		dto.ResetPasswordRequest	true	"Email for password reset"
//	@Success		200		{object}	dto.SuccessResponse			"If the email exists, reset instructions have been sent"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/user/reset_password [post]
func (s *userService) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "Invalid request body")
		return
	}

	ctx := c.Request.Context()
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusOK, dto.SuccessResponse{Message: "If the email exists, reset instructions have been sent"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, ResetPasswordClaims{
		UserID:         user.ID,
		OrganizationID: user.OrganizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	tokenString, err := token.SignedString([]byte(s.cfg.SecretKeyJWT + "reset_password"))
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to generate reset token")
		return
	}

	resetLink := fmt.Sprintf("%s/api/v1/user/reset_password/confirm?token=%s", s.cfg.Protocol+"://"+s.cfg.Host+":"+s.cfg.Port, tokenString)
	if err := s.emailService.SendPasswordResetEmail(user.Email, user.Name, resetLink); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to send reset email")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "If the email exists, reset instructions have been sent"})
}

// ConfirmResetPasswordHandler godoc
//
//	@Summary		Confirm password reset
//	@Description	Accepts new password from user using a valid reset token. Redirects user to onboarding with JWT.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			token	query		string							true	"Password reset token from email link"
//	@Param			data	body		dto.ConfirmResetPasswordRequest	true	"New password"
//	@Success		302		{string}	string							"Redirects to /onboarding with accessToken and refreshToken"
//	@Failure		400		{object}	lib.ErrorResponse				"Bad request / invalid token / missing password"
//	@Failure		401		{object}	dto.UnauthorizedResponse		"Unauthorized / expired token"
//	@Failure		500		{object}	lib.ErrorResponse				"Internal server error"
//	@Router			/api/v1/user/reset_password/confirm [get]
func (s *userService) ConfirmResetPasswordHandler(c *gin.Context) {
	telemetry.Log.Info("ConfirmResetPasswordHandler")
	tokenStr := c.Query("token")
	if tokenStr == "" {
		lib.ResponseBadRequest(c, errors.New("missing token"), "Missing token")
		return
	}

	token, err := jwt.ParseWithClaims(tokenStr, &ResetPasswordClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.SecretKeyJWT + "reset_password"), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "Invalid or expired token"})
		return
	}

	claims := token.Claims.(*ResetPasswordClaims)
	userID := claims.UserID

	ctx := c.Request.Context()
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		lib.ResponseBadRequest(c, errors.New("user not found"), "User does not exist")
		return
	}

	user.Password = ""
	user.Status = model.UserStatuses.ACTIVE
	user.Verified = true

	if err := s.userRepo.UpdateUser(ctx, user.ID, *user); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to update user password")
		return
	}

	jwtToken, err := s.jwtAuth.TokenGenerator(user)
	if err != nil {
		telemetry.Log.Error(err.Error())
		c.Redirect(http.StatusFound, s.cfg.FrontendUrl+"/onboarding?error=internal%20error%20token")
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("%s/onboarding?accessToken=%s&refreshToken=%s", s.cfg.FrontendUrl, jwtToken.AccessToken, jwtToken.RefreshToken))
}
