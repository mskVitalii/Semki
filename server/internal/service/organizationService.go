package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/internal/utils/jwtUtils"
	"semki/internal/utils/mongoUtils"
	"semki/pkg/lib"
	"semki/pkg/telemetry"
	"strconv"
)

// organizationService - dependent services
type organizationService struct {
	orgRepo  mongo.IOrganizationRepository
	userRepo mongo.IUserRepository
}

func NewOrganizationService(orgRepo mongo.IOrganizationRepository, userRepo mongo.IUserRepository) routes.IOrganizationService {
	return &organizationService{orgRepo, userRepo}
}

// CreateOrganization godoc
//
//	@Summary		Creates a new organization
//	@Description	Creates a new organization in the MongoDB database.
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			organization	body		dto.CreateOrganizationRequest	true	"Organization object to create"
//	@Success		201				{object}	dto.CreateOrganizationResponse	"Successful response"
//	@Failure		400				{object}	lib.ErrorResponse				"Bad request"
//	@Failure		500				{object}	lib.ErrorResponse				"Internal server error"
//	@Router			/api/v1/organization [post]
func (s *organizationService) CreateOrganization(c *gin.Context) {
	var organizationDto dto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&organizationDto); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	ctx := c.Request.Context()

	organizationByTitle, err := s.orgRepo.GetOrganizationByTitle(ctx, organizationDto.Title)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to check organization existence")
		return
	}
	if organizationByTitle != nil {
		lib.ResponseBadRequest(c, errors.New("Organization already exists"), "Organization already exists")
		return
	}

	// Creating organization
	organization := dto.NewOrganizationFromRequest(organizationDto)

	if err := s.orgRepo.CreateOrganization(ctx, organization); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create organization")
		return
	}

	c.JSON(http.StatusCreated, dto.CreateOrganizationResponse{Message: "Organization created"})
}

// GetOrganization godoc
//
//	@Summary		Retrieves an organization by user claims
//	@Description	Retrieves an organization from the MongoDB database by its ID.
//	@Tags			organizations
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GetOrganizationResponse	"Successful response"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		404	{object}	lib.ErrorResponse			"Organization not found"
//	@Failure		500	{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization [get]
func (s *organizationService) GetOrganization(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId
	telemetry.Log.Info(fmt.Sprintf("GetOrganization -> organizationId%s", organizationId))

	ctx := c.Request.Context()
	organization, err := s.orgRepo.GetOrganizationByID(ctx, organizationId)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to get organization")
		return
	}
	telemetry.Log.Info(fmt.Sprintf("GetOrganization -> organization is nil %t", organization == nil))

	if organization == nil {
		lib.ResponseNotFound(c, "Organization not found")
		return
	}
	telemetry.Log.Info(fmt.Sprintf("GetOrganization -> organization title %s", organization.Title))

	c.JSON(http.StatusOK, organization)
}

// GetOrganizationUsers godoc
//
//	@Summary		Retrieves paginated organization users
//	@Description	Retrieves users of the current user's organization with optional search
//	@Tags			organizations
//	@Produce		json
//	@Param			page	query	int		false	"Page number"				default(1)
//	@Param			limit	query	int		false	"Number of users per page"	default(20)
//	@Param			search	query	string	false	"Search by name or email"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GetOrganizationUsersResponse	"Successful response"
//	@Failure		401	{object}	dto.UnauthorizedResponse			"Unauthorized"
//	@Failure		404	{object}	lib.ErrorResponse					"Organization not found"
//	@Failure		500	{object}	lib.ErrorResponse					"Internal server error"
//	@Router			/api/v1/organization/users [get]
func (s *organizationService) GetOrganizationUsers(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	orgID := userClaims.(*jwtUtils.UserClaims).OrganizationId

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	search := c.Query("search")

	page := 1
	limit := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := c.Request.Context()
	users, totalCount, err := s.userRepo.GetUsersByOrganization(ctx, orgID, search, page, limit)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to get organization users")
		return
	}

	c.JSON(http.StatusOK, dto.GetOrganizationUsersResponse{
		Users:      users,
		TotalCount: totalCount,
	})
}

// UpdateOrganization godoc
//
//	@Summary		Updates an organization by its ID
//	@Description	Updates an organization in the MongoDB database by its ID.
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			organization	body		model.Organization				true	"Organization object with updated data"
//	@Success		200				{object}	dto.UpdateOrganizationResponse	"Successful response"
//	@Failure		401				{object}	dto.UnauthorizedResponse		"Unauthorized"
//	@Failure		400				{object}	lib.ErrorResponse				"Bad request"
//	@Failure		500				{object}	lib.ErrorResponse				"Internal server error"
//	@Router			/api/v1/organization/{id} [put]
//func (s *organizationService) UpdateOrganization(c *gin.Context) {
//	id := c.Param("id")
//	paramObjectId, err := mongoUtils.StringToObjectID(id)
//	if err != nil {
//		lib.ResponseBadRequest(c, errors.New("wrong organization id"), "Wrong id format")
//		return
//	}
//
//	var organization model.Organization
//	if err := c.ShouldBindJSON(&organization); err != nil {
//		lib.ResponseBadRequest(c, err, "Failed to bind body")
//		return
//	}
//
//	userClaims, _ := c.Get(jwtUtils.IdentityKey)
//	if userClaims == nil {
//		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
//		return
//	}
//	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId
//
//	if organizationId != organization.ID || organizationId != paramObjectId {
//		lib.ResponseBadRequest(c, errors.New("Wrong organization id"), "Organization id must be the same organization")
//		return
//	}
//	ctx := c.Request.Context()
//
//	if err := s.chatRepo.UpdateOrganization(ctx, paramObjectId, organization); err != nil {
//		lib.ResponseInternalServerError(c, err, "Failed to update organization")
//		return
//	}
//
//	c.JSON(http.StatusOK, dto.UpdateOrganizationResponse{Message: "Organization updated"})
//}

// DeleteOrganization godoc
//
//	@Summary		Deletes an organization by its ID
//	@Description	Deletes an organization from the MongoDB database by its ID.
//	@Tags			organizations
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.DeleteOrganizationResponse	"Successful response"
//	@Failure		401	{object}	dto.UnauthorizedResponse		"Unauthorized"
//	@Failure		500	{object}	lib.ErrorResponse				"Internal server error"
//	@Router			/api/v1/organization [delete]
func (s *organizationService) DeleteOrganization(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	claims, ok := userClaims.(*jwtUtils.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "No claims"})
		return
	}
	organizationId := claims.OrganizationId
	if claims.OrganizationRole != model.OrganizationRoles.ADMIN || claims.OrganizationRole != model.OrganizationRoles.OWNER {
		lib.ResponseBadRequest(c, errors.New("Not enough rights"), "Not enough rights")
		return
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.DeleteOrganization(ctx, organizationId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to delete organization")
		return
	}

	c.JSON(http.StatusOK, dto.DeleteOrganizationResponse{Message: "Organization deleted"})
}

// PatchOrganization godoc
//
//	@Summary		Partially updates an organization
//	@Description	Updates only the provided fields of an organization
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			organization	body		dto.PatchOrganizationRequest	true	"Fields to update"
//	@Success		200				{object}	dto.PatchOrganizationResponse	"Successful response"
//	@Failure		401				{object}	dto.UnauthorizedResponse		"Unauthorized"
//	@Failure		400				{object}	lib.ErrorResponse				"Bad request"
//	@Failure		500				{object}	lib.ErrorResponse				"Internal server error"
//	@Router			/api/v1/organization [patch]
func (s *organizationService) PatchOrganization(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	var req dto.PatchOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	updates := bson.M{}
	if req.Title != "" {
		updates["title"] = req.Title
	}

	if len(updates) == 0 {
		lib.ResponseBadRequest(c, errors.New("no fields to update"), "No fields to update")
		return
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.PatchOrganization(ctx, organizationId, updates); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to update organization")
		return
	}

	c.JSON(http.StatusOK, dto.PatchOrganizationResponse{Message: "Organization updated"})
}

// CreateTeam godoc
//
//	@Summary		Adds a new team to organization
//	@Description	Creates a new team in the organization's semantic structure
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team	body		dto.CreateTeamRequest		true	"Team data"
//	@Success		201		{object}	dto.TeamResponse			"Successful response"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/teams [post]
func (s *organizationService) CreateTeam(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	var req dto.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	team := model.Team{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Description: req.Description,
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.AddTeam(ctx, organizationId, team); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create team")
		return
	}

	c.JSON(http.StatusCreated, dto.TeamResponse{Message: "Team created"})
}

// UpdateTeam godoc
//
//	@Summary		Updates a team
//	@Description	Updates team information by ID
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			teamId	path		string						true	"Team ID"
//	@Param			team	body		dto.UpdateTeamRequest		true	"Team data to update"
//	@Success		200		{object}	dto.TeamResponse			"Successful response"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/teams/{teamId} [put]
func (s *organizationService) UpdateTeam(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	teamIdStr := c.Param("teamId")
	teamId, err := mongoUtils.StringToObjectID(teamIdStr)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong team id"), "Wrong team id format")
		return
	}

	var req dto.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	updates := bson.M{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		lib.ResponseBadRequest(c, errors.New("no fields to update"), "No fields to update")
		return
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.UpdateTeam(ctx, organizationId, teamId, updates); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to update team")
		return
	}

	c.JSON(http.StatusOK, dto.TeamResponse{Message: "Team updated"})
}

// DeleteTeam godoc
//
//	@Summary		Deletes a team
//	@Description	Removes a team from organization by ID
//	@Tags			organizations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			teamId	path		string						true	"Team ID"
//	@Success		200		{object}	dto.TeamResponse			"Successful response"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/teams/{teamId} [delete]
func (s *organizationService) DeleteTeam(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	teamIdStr := c.Param("teamId")
	teamId, err := mongoUtils.StringToObjectID(teamIdStr)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong team id"), "Wrong team id format")
		return
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.DeleteTeam(ctx, organizationId, teamId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to delete team")
		return
	}

	c.JSON(http.StatusOK, dto.TeamResponse{Message: "Team deleted"})
}

// CreateLevel godoc
//
//	@Summary		Adds a new level to organization
//	@Description	Creates a new level in the organization's semantic structure
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			level	body		dto.CreateLevelRequest		true	"Level data"
//	@Success		201		{object}	dto.LevelResponse			"Successful response"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/levels [post]
func (s *organizationService) CreateLevel(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	var req dto.CreateLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	level := model.Level{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Description: req.Description,
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.AddLevel(ctx, organizationId, level); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create level")
		return
	}

	c.JSON(http.StatusCreated, dto.LevelResponse{Message: "Level created"})
}

// UpdateLevel godoc
//
//	@Summary		Updates a level
//	@Description	Updates level information by ID
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			levelId	path		string						true	"Level ID"
//	@Param			level	body		dto.UpdateLevelRequest		true	"Level data to update"
//	@Success		200		{object}	dto.LevelResponse			"Successful response"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/levels/{levelId} [put]
func (s *organizationService) UpdateLevel(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	levelIdStr := c.Param("levelId")
	levelId, err := mongoUtils.StringToObjectID(levelIdStr)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong level id"), "Wrong level id format")
		return
	}

	var req dto.UpdateLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	updates := bson.M{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		lib.ResponseBadRequest(c, errors.New("no fields to update"), "No fields to update")
		return
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.UpdateLevel(ctx, organizationId, levelId, updates); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to update level")
		return
	}

	c.JSON(http.StatusOK, dto.LevelResponse{Message: "Level updated"})
}

// DeleteLevel godoc
//
//	@Summary		Deletes a level
//	@Description	Removes a level from organization by ID
//	@Tags			organizations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			levelId	path		string						true	"Level ID"
//	@Success		200		{object}	dto.LevelResponse			"Successful response"
//	@Failure		401		{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400		{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500		{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/levels/{levelId} [delete]
func (s *organizationService) DeleteLevel(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	levelIdStr := c.Param("levelId")
	levelId, err := mongoUtils.StringToObjectID(levelIdStr)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong level id"), "Wrong level id format")
		return
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.DeleteLevel(ctx, organizationId, levelId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to delete level")
		return
	}

	c.JSON(http.StatusOK, dto.LevelResponse{Message: "Level deleted"})
}

// CreateLocation godoc
//
//	@Summary		Adds a new location to organization
//	@Description	Creates a new location in the organization's semantic structure
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			location	body		dto.CreateLocationRequest	true	"Location data"
//	@Success		201			{object}	dto.LocationResponse		"Successful response"
//	@Failure		401			{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400			{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500			{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/locations [post]
func (s *organizationService) CreateLocation(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	var req dto.CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	location := model.Location{
		ID:   primitive.NewObjectID(),
		Name: req.Name,
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.AddLocation(ctx, organizationId, location); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create location")
		return
	}

	c.JSON(http.StatusCreated, dto.LocationResponse{Message: "Location created"})
}

// DeleteLocation godoc
//
//	@Summary		Deletes a location
//	@Description	Removes a location from organization by ID
//	@Tags			organizations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			locationId	path		string						true	"Location ID"
//	@Success		200			{object}	dto.LocationResponse		"Successful response"
//	@Failure		401			{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		400			{object}	lib.ErrorResponse			"Bad request"
//	@Failure		500			{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/locations/{locationId} [delete]
func (s *organizationService) DeleteLocation(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	locationIdStr := c.Param("locationId")
	locationId, err := mongoUtils.StringToObjectID(locationIdStr)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong location id"), "Wrong location id format")
		return
	}

	ctx := c.Request.Context()
	if err := s.orgRepo.DeleteLocation(ctx, organizationId, locationId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to delete location")
		return
	}

	c.JSON(http.StatusOK, dto.LocationResponse{Message: "Location deleted"})
}
