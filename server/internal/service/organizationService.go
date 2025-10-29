package service

import (
	"fmt"
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
	"strconv"
)

// organizationService - dependent services
type organizationService struct {
	repo mongo.IRepository
}

func NewOrganizationService(repo mongo.IRepository) routes.IOrganizationService {
	return &organizationService{repo}
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

	organizationByTitle, err := s.repo.GetOrganizationByTitle(ctx, organizationDto.Title)
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

	if err := s.repo.CreateOrganization(ctx, organization); err != nil {
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
	organization, err := s.repo.GetOrganizationByID(ctx, organizationId)
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
	users, totalCount, err := s.repo.GetUsersByOrganization(ctx, orgID, search, page, limit)
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
func (s *organizationService) UpdateOrganization(c *gin.Context) {
	id := c.Param("id")
	paramObjectId, err := mongoUtils.StringToObjectID(id)
	if err != nil {
		lib.ResponseBadRequest(c, errors.New("wrong organization id"), "Wrong id format")
		return
	}

	var organization model.Organization
	if err := c.ShouldBindJSON(&organization); err != nil {
		lib.ResponseBadRequest(c, err, "Failed to bind body")
		return
	}

	userClaims, _ := c.Get(jwtUtils.IdentityKey)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Message: "unauthorized"})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId

	if organizationId != organization.ID || organizationId != paramObjectId {
		lib.ResponseBadRequest(c, errors.New("Wrong organization id"), "Organization id must be the same organization")
		return
	}
	ctx := c.Request.Context()

	if err := s.repo.UpdateOrganization(ctx, paramObjectId, organization); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to update organization")
		return
	}

	c.JSON(http.StatusOK, dto.UpdateOrganizationResponse{Message: "Organization updated"})
}

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
	if err := s.repo.DeleteOrganization(ctx, organizationId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to delete organization")
		return
	}

	c.JSON(http.StatusOK, dto.DeleteOrganizationResponse{Message: "Organization deleted"})
}
