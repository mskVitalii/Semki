package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
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
	"strings"
)

// organizationService - dependent services
type organizationService struct {
	orgRepo       mongo.IOrganizationRepository
	userRepo      mongo.IUserRepository
	qdrantService IQdrantService
}

func NewOrganizationService(orgRepo mongo.IOrganizationRepository, userRepo mongo.IUserRepository, qdrantService IQdrantService) routes.IOrganizationService {
	return &organizationService{orgRepo, userRepo, qdrantService}
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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID
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
	orgID := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
//	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID
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
	organizationId := claims.OrganizationID
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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationID

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

// InsertMock godoc
//
//	@Summary		Inserts mock data into organization
//	@Description	Populates the organization with sample teams, levels, and locations for testing
//	@Tags			organizations
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.LocationResponse		"Successful response"
//	@Failure		401	{object}	dto.UnauthorizedResponse	"Unauthorized"
//	@Failure		500	{object}	lib.ErrorResponse			"Internal server error"
//	@Router			/api/v1/organization/insert-mock [post]
func (s *organizationService) InsertMock(c *gin.Context) {
	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationID := userClaims.(*jwtUtils.UserClaims).OrganizationID

	ctx := c.Request.Context()
	org, err := s.orgRepo.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to load organization data")
		return
	}

	existingTeams := make(map[string]model.Team)
	for _, t := range org.Semantic.Teams {
		existingTeams[t.Name] = t
	}

	existingLevels := make(map[string]model.Level)
	for _, l := range org.Semantic.Levels {
		existingLevels[l.Name] = l
	}

	existingLocations := make(map[string]model.Location)
	for _, loc := range org.Semantic.Locations {
		existingLocations[loc.Name] = loc
	}
	mockOrgEmailSuffix := "@" + strings.ReplaceAll(strings.ToLower(org.Title), " ", " ") + ".com"
	// Mock Teams
	mockTeams := []model.Team{
		{
			ID:          primitive.NewObjectID(),
			Name:        "Engineering",
			Description: "Builds and maintains software systems, APIs, and infrastructure. Ask them about technical feasibility, performance, and implementation details. Do not ask about marketing strategy or product positioning.",
		},
		{
			ID:          primitive.NewObjectID(),
			Name:        "Product",
			Description: "Defines product vision, roadmap, and priorities. Ask them about goals, requirements, and business value. Do not ask about code-level or visual design decisions.",
		},
		{
			ID:          primitive.NewObjectID(),
			Name:        "Design",
			Description: "Focuses on UX, UI, and user research. Ask them about usability, visual consistency, and user flows. Do not ask about backend logic or business strategy.",
		},
		{
			ID:          primitive.NewObjectID(),
			Name:        "Marketing",
			Description: "Handles brand awareness, campaigns, and user acquisition. Ask them about messaging, audience targeting, and market trends. Do not ask about product functionality or engineering details.",
		},
		{
			ID:          primitive.NewObjectID(),
			Name:        "Sales",
			Description: "Drives revenue through client relations and deals. Ask them about customer needs, pricing, and contracts. Do not ask about design, implementation, or marketing content.",
		},
	}

	// Mock Levels
	mockLevels := []model.Level{
		{
			ID:          primitive.NewObjectID(),
			Name:        "Junior",
			Description: "Entry-level engineer (0–2 years). Handles simple, well-defined tasks under supervision. Ask them about small fixes, documentation, or learning tasks. Do not ask about architecture, estimations, or code reviews.",
		},
		{
			ID:          primitive.NewObjectID(),
			Name:        "Middle",
			Description: "Mid-level engineer (2–5 years). Works independently and delivers complete features. Ask them about implementation details, tools, and best practices. Do not ask about architecture or long-term decisions.",
		},
		{
			ID:          primitive.NewObjectID(),
			Name:        "Senior",
			Description: "Experienced engineer (5+ years). Designs solutions, mentors others, and ensures code quality. Ask them about architecture, optimization, and complex issues. Do not ask about trivial or routine tasks.",
		},
		{
			ID:          primitive.NewObjectID(),
			Name:        "Lead",
			Description: "Team leader responsible for planning, priorities, and communication. Ask them about task alignment, blockers, or coordination across teams. Do not ask about minor technical issues.",
		},
		{
			ID:          primitive.NewObjectID(),
			Name:        "Principal",
			Description: "Principal engineer defining technical vision and long-term strategy. Ask them about architecture standards, technology choices, and product direction. Do not ask about operational or feature-level tasks.",
		},
	}

	// Mock Locations
	mockLocations := []model.Location{
		{
			ID:   primitive.NewObjectID(),
			Name: "San Francisco, CA",
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "New York, NY",
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "London, UK",
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "Berlin, Germany",
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "Tokyo, Japan",
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "Remote",
		},
	}

	for i, team := range mockTeams {
		if existing, ok := existingTeams[team.Name]; ok {
			mockTeams[i] = existing
			continue
		}
		if err := s.orgRepo.AddTeam(ctx, organizationID, team); err != nil {
			telemetry.Log.Warn("Failed to insert mock team: "+team.Name, zap.Error(err))
		}
	}

	for i, level := range mockLevels {
		if existing, ok := existingLevels[level.Name]; ok {
			mockLevels[i] = existing
			continue
		}
		if err := s.orgRepo.AddLevel(ctx, organizationID, level); err != nil {
			telemetry.Log.Warn("Failed to insert mock level: "+level.Name, zap.Error(err))
		}
	}

	for i, location := range mockLocations {
		if existing, ok := existingLocations[location.Name]; ok {
			mockLocations[i] = existing
			continue
		}
		if err := s.orgRepo.AddLocation(ctx, organizationID, location); err != nil {
			telemetry.Log.Warn("Failed to insert mock location: "+location.Name, zap.Error(err))
		}
	}

	// Generate mock users only if we have teams, levels, and locations
	mockUsersCount := 0

	mockUsers := []model.User{
		{
			ID:               primitive.NewObjectID(),
			Email:            "alice.chen" + mockOrgEmailSuffix,
			Name:             "Alice Chen",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Senior full-stack engineer with expertise in React and Go. Known for writing clean, maintainable code and mentoring junior developers. Prefers async communication and detailed technical documentation.",
				Team:        mockTeams[0].ID,     // Engineering
				Level:       mockLevels[2].ID,    // Senior
				Location:    mockLocations[0].ID, // San Francisco
			},
			Contact: model.UserContact{
				Email: "alice.chen" + mockOrgEmailSuffix,
				Slack: "@alice",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "bob.martinez" + mockOrgEmailSuffix,
			Name:             "Bob Martinez",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Product manager focused on user experience and data-driven decisions. Excellent at stakeholder management and roadmap planning. Responds quickly to Slack messages during business hours.",
				Team:        mockTeams[1].ID,     // Product
				Level:       mockLevels[1].ID,    // Middle
				Location:    mockLocations[1].ID, // New York
			},
			Contact: model.UserContact{
				Email: "bob.martinez" + mockOrgEmailSuffix,
				Slack: "@bob",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "carol.wang" + mockOrgEmailSuffix,
			Name:             "Carol Wang",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Principal UX designer specializing in design systems and accessibility. Strong advocate for user research and iterative design. Prefers visual communication through Figma comments and screenshots.",
				Team:        mockTeams[2].ID,     // Design
				Level:       mockLevels[4].ID,    // Principal
				Location:    mockLocations[2].ID, // London
			},
			Contact: model.UserContact{
				Email: "carol.wang" + mockOrgEmailSuffix,
				Slack: "@carol",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "david.kowalski" + mockOrgEmailSuffix,
			Name:             "David Kowalski",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Marketing lead with deep expertise in growth strategies and content marketing. Data-focused and experienced in A/B testing. Works across multiple time zones and prefers email for formal communications.",
				Team:        mockTeams[3].ID,     // Marketing
				Level:       mockLevels[3].ID,    // Lead
				Location:    mockLocations[3].ID, // Berlin
			},
			Contact: model.UserContact{
				Email: "david.kowalski" + mockOrgEmailSuffix,
				Slack: "@david",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "emma.jackson" + mockOrgEmailSuffix,
			Name:             "Emma Jackson",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Junior sales representative with strong interpersonal skills and enthusiasm for learning. Great at building relationships with clients. Prefers phone calls and video meetings over text-based communication.",
				Team:        mockTeams[4].ID,     // Sales
				Level:       mockLevels[0].ID,    // Junior
				Location:    mockLocations[4].ID, // Tokyo
			},
			Contact: model.UserContact{
				Email: "emma.jackson" + mockOrgEmailSuffix,
				Slack: "@emma",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "frank.oconnor" + mockOrgEmailSuffix,
			Name:             "Frank O'Connor",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Backend engineer specializing in microservices and database optimization. Self-directed worker who prefers written specs and minimal meetings. Highly responsive to code review requests.",
				Team:        mockTeams[0].ID,     // Engineering
				Level:       mockLevels[1].ID,    // Middle
				Location:    mockLocations[5].ID, // Remote
			},
			Contact: model.UserContact{
				Email: "frank.oconnor" + mockOrgEmailSuffix,
				Slack: "@frank",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "grace.thompson" + mockOrgEmailSuffix,
			Name:             "Grace Thompson",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Product designer with focus on mobile interfaces and interaction design. Detail-oriented and passionate about pixel-perfect implementations. Collaborates best with real-time design feedback sessions.",
				Team:        mockTeams[2].ID,     // Design
				Level:       mockLevels[1].ID,    // Middle
				Location:    mockLocations[5].ID, // Remote
			},
			Contact: model.UserContact{
				Email: "grace.thompson" + mockOrgEmailSuffix,
				Slack: "@grace",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "henry.kim" + mockOrgEmailSuffix,
			Name:             "Henry Kim",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Senior sales executive with expertise in enterprise deals and contract negotiations. Strategic thinker with strong business acumen. Available via phone during business hours, checks email frequently.",
				Team:        mockTeams[4].ID,     // Sales
				Level:       mockLevels[2].ID,    // Senior
				Location:    mockLocations[1].ID, // New York
			},
			Contact: model.UserContact{
				Email: "henry.kim" + mockOrgEmailSuffix,
				Slack: "@henry",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "isabella.nguyen" + mockOrgEmailSuffix,
			Name:             "Isabella Nguyen",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.ADMIN,
			Semantic: model.UserSemantic{
				Description: "Engineering manager overseeing full-stack teams. Skilled in agile methodologies, performance reviews, and technical strategy. Prefers asynchronous updates but joins weekly syncs.",
				Team:        mockTeams[0].ID,  // Engineering
				Level:       mockLevels[3].ID, // Lead
				Location:    mockLocations[1].ID,
			},
			Contact: model.UserContact{
				Email: "isabella.nguyen" + mockOrgEmailSuffix,
				Slack: "@isabella",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "jack.tanaka" + mockOrgEmailSuffix,
			Name:             "Jack Tanaka",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Data analyst experienced in SQL, Python, and BI tools. Provides actionable insights to marketing and product teams. Prefers structured documentation and concise reports.",
				Team:        mockTeams[3].ID,  // Marketing
				Level:       mockLevels[1].ID, // Middle
				Location:    mockLocations[4].ID,
			},
			Contact: model.UserContact{
				Email: "jack.tanaka" + mockOrgEmailSuffix,
				Slack: "@jack",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "karen.meyer" + mockOrgEmailSuffix,
			Name:             "Karen Meyer",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.ADMIN,
			Semantic: model.UserSemantic{
				Description: "Head of Product leading roadmap execution and cross-functional coordination. Strong communicator with focus on strategic outcomes. Prefers brief updates via email.",
				Team:        mockTeams[1].ID,  // Product
				Level:       mockLevels[4].ID, // Principal
				Location:    mockLocations[0].ID,
			},
			Contact: model.UserContact{
				Email: "karen.meyer" + mockOrgEmailSuffix,
				Slack: "@karen",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "liam.santos" + mockOrgEmailSuffix,
			Name:             "Liam Santos",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Frontend developer passionate about performance optimization and accessibility. Enjoys pairing sessions and clear design specs.",
				Team:        mockTeams[0].ID,  // Engineering
				Level:       mockLevels[1].ID, // Middle
				Location:    mockLocations[3].ID,
			},
			Contact: model.UserContact{
				Email: "liam.santos" + mockOrgEmailSuffix,
				Slack: "@liam",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "mia.garcia" + mockOrgEmailSuffix,
			Name:             "Mia Garcia",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "QA engineer ensuring product quality through automated testing. Strong believer in CI/CD best practices and detailed bug reports.",
				Team:        mockTeams[0].ID,  // Engineering
				Level:       mockLevels[1].ID, // Middle
				Location:    mockLocations[2].ID,
			},
			Contact: model.UserContact{
				Email: "mia.garcia" + mockOrgEmailSuffix,
				Slack: "@mia",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "noah.schneider" + mockOrgEmailSuffix,
			Name:             "Noah Schneider",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "DevOps specialist managing cloud infrastructure and deployment pipelines. Prefers infrastructure-as-code and well-defined alerting policies.",
				Team:        mockTeams[0].ID,  // Engineering
				Level:       mockLevels[2].ID, // Senior
				Location:    mockLocations[5].ID,
			},
			Contact: model.UserContact{
				Email: "noah.schneider" + mockOrgEmailSuffix,
				Slack: "@noah",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "olivia.ross" + mockOrgEmailSuffix,
			Name:             "Olivia Ross",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Marketing analyst passionate about campaign optimization and lead scoring. Excellent at presenting data insights to non-technical audiences.",
				Team:        mockTeams[3].ID,  // Marketing
				Level:       mockLevels[1].ID, // Middle
				Location:    mockLocations[1].ID,
			},
			Contact: model.UserContact{
				Email: "olivia.ross" + mockOrgEmailSuffix,
				Slack: "@olivia",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "patrick.larsen" + mockOrgEmailSuffix,
			Name:             "Patrick Larsen",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Technical writer creating documentation for developers and end-users. Values clarity, precision, and feedback loops from engineers.",
				Team:        mockTeams[1].ID,  // Product
				Level:       mockLevels[2].ID, // Senior
				Location:    mockLocations[5].ID,
			},
			Contact: model.UserContact{
				Email: "patrick.larsen" + mockOrgEmailSuffix,
				Slack: "@patrick",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "quinn.hart" + mockOrgEmailSuffix,
			Name:             "Quinn Hart",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "Support specialist providing customer solutions with empathy and precision. Expert in CRM systems and escalation management.",
				Team:        mockTeams[4].ID,  // Sales
				Level:       mockLevels[1].ID, // Middle
				Location:    mockLocations[3].ID,
			},
			Contact: model.UserContact{
				Email: "quinn.hart" + mockOrgEmailSuffix,
				Slack: "@quinn",
			},
		},
		{
			ID:               primitive.NewObjectID(),
			Email:            "rachel.kowalska" + mockOrgEmailSuffix,
			Name:             "Rachel Kowalska",
			Providers:        []model.UserProvider{model.UserProviders.Email},
			Verified:         true,
			Status:           model.UserStatuses.ACTIVE,
			OrganizationID:   organizationID,
			OrganizationRole: model.OrganizationRoles.USER,
			Semantic: model.UserSemantic{
				Description: "HR specialist managing hiring pipelines and onboarding. Skilled in negotiation, performance evaluation, and employee engagement programs.",
				Team:        mockTeams[4].ID,  // Sales (HR adjacent)
				Level:       mockLevels[2].ID, // Senior
				Location:    mockLocations[0].ID,
			},
			Contact: model.UserContact{
				Email: "rachel.kowalska" + mockOrgEmailSuffix,
				Slack: "@rachel",
			},
		},
	}

	// Insert users
	for _, user := range mockUsers {
		if err := s.userRepo.CreateUser(ctx, &user); err != nil {
			telemetry.Log.Warn("Failed to insert mock user: "+user.Name, zap.Error(err))
		} else {
			mockUsersCount++
		}
	}
	_, err = s.qdrantService.ReIndexFunc(ctx, organizationID)
	if err != nil {
		lib.ResponseInternalServerError(c, err, "failed to fetch users")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mock data inserted successfully",
		"data": gin.H{
			"teams":     len(mockTeams),
			"levels":    len(mockLevels),
			"locations": len(mockLocations),
			"users":     mockUsersCount,
		},
	})
}
