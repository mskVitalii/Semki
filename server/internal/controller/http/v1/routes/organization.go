package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	organizationCRUD = "/organization"
)

type IOrganizationService interface {
	CreateOrganization(c *gin.Context)
	GetOrganization(c *gin.Context)
	UpdateOrganization(c *gin.Context)
	DeleteOrganization(c *gin.Context)
	GetOrganizationUsers(c *gin.Context)
}

func RegisterOrganizationRoutes(g *gin.RouterGroup, organizationService IOrganizationService, securityHandler gin.HandlerFunc) {
	g.POST(organizationCRUD, organizationService.CreateOrganization)
	g.GET(organizationCRUD, securityHandler, organizationService.GetOrganization)
	g.GET(organizationCRUD+"/users", securityHandler, organizationService.GetOrganizationUsers)
	g.PUT(organizationCRUD, securityHandler, organizationService.UpdateOrganization)
	g.DELETE(organizationCRUD, securityHandler, organizationService.DeleteOrganization)
}
