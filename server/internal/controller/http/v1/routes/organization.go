package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	organizationCRUD      = "/organization"
	organizationTeams     = "/organization/teams"
	organizationLevels    = "/organization/levels"
	organizationLocations = "/organization/locations"
)

type IOrganizationService interface {
	CreateOrganization(c *gin.Context)
	GetOrganization(c *gin.Context)
	DeleteOrganization(c *gin.Context)
	GetOrganizationUsers(c *gin.Context)
	PatchOrganization(c *gin.Context)

	CreateTeam(c *gin.Context)
	UpdateTeam(c *gin.Context)
	DeleteTeam(c *gin.Context)

	CreateLevel(c *gin.Context)
	UpdateLevel(c *gin.Context)
	DeleteLevel(c *gin.Context)

	CreateLocation(c *gin.Context)
	DeleteLocation(c *gin.Context)

	//UpdateOrganization(c *gin.Context)

	InsertMock(c *gin.Context)
}

func RegisterOrganizationRoutes(g *gin.RouterGroup, organizationService IOrganizationService, securityHandler gin.HandlerFunc) {
	g.POST(organizationCRUD, organizationService.CreateOrganization)
	g.GET(organizationCRUD, securityHandler, organizationService.GetOrganization)
	g.GET(organizationCRUD+"/users", securityHandler, organizationService.GetOrganizationUsers)
	//g.PUT(organizationCRUD, securityHandler, organizationService.UpdateOrganization)
	g.DELETE(organizationCRUD, securityHandler, organizationService.DeleteOrganization)
	g.PATCH(organizationCRUD, securityHandler, organizationService.PatchOrganization)

	// Teams
	g.POST(organizationTeams, securityHandler, organizationService.CreateTeam)
	g.PUT(organizationTeams+"/:teamId", securityHandler, organizationService.UpdateTeam)
	g.DELETE(organizationTeams+"/:teamId", securityHandler, organizationService.DeleteTeam)

	// Levels
	g.POST(organizationLevels, securityHandler, organizationService.CreateLevel)
	g.PUT(organizationLevels+"/:levelId", securityHandler, organizationService.UpdateLevel)
	g.DELETE(organizationLevels+"/:levelId", securityHandler, organizationService.DeleteLevel)

	// Locations
	g.POST(organizationLocations, securityHandler, organizationService.CreateLocation)
	g.DELETE(organizationLocations+"/:locationId", securityHandler, organizationService.DeleteLocation)

	g.POST(organizationCRUD+"/insert-mock", securityHandler, organizationService.InsertMock)
}
