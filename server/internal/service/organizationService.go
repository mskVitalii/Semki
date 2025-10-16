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
)

// organizationService - dependent services
type organizationService struct {
	mongoRepo mongo.IMongoRepository
}

func NewOrganizationService(mongoRepo mongo.IMongoRepository) routes.IOrganizationService {
	return &organizationService{mongoRepo}
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

	organizationByTitle, err := s.mongoRepo.GetOrganizationByTitle(ctx, organizationDto.Title)
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

	if err := s.mongoRepo.CreateOrganization(ctx, organization); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to create organization")
		return
	}

	c.JSON(http.StatusCreated, dto.CreateOrganizationResponse{Message: "Organization created"})
}

// GetOrganization godoc
//
//	@Summary		Retrieves an organization by its ID
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
		c.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{
			Message: "unauthorized",
		})
		return
	}
	organizationId := userClaims.(*jwtUtils.UserClaims).OrganizationId
	telemetry.Log.Info(fmt.Sprintf("GetOrganization -> organizationId%s", organizationId))

	ctx := c.Request.Context()
	organization, err := s.mongoRepo.GetOrganizationByID(ctx, organizationId)
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

	if organizationId != organization.Id || organizationId != paramObjectId {
		lib.ResponseBadRequest(c, errors.New("Wrong organization id"), "Organization id must be the same organization")
		return
	}
	ctx := c.Request.Context()

	if err := s.mongoRepo.UpdateOrganization(ctx, paramObjectId, organization); err != nil {
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
	if err := s.mongoRepo.DeleteOrganization(ctx, organizationId); err != nil {
		lib.ResponseInternalServerError(c, err, "Failed to delete organization")
		return
	}

	c.JSON(http.StatusOK, dto.DeleteOrganizationResponse{Message: "Organization deleted"})
}
