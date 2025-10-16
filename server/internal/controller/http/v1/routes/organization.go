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
}

func RegisterOrganizationRoutes(g *gin.RouterGroup, organizationService IOrganizationService, securityHandler gin.HandlerFunc) {
	g.POST(organizationCRUD, organizationService.CreateOrganization)
	g.GET(organizationCRUD+"/:id", securityHandler, organizationService.GetOrganization)
	g.PUT(organizationCRUD+"/:id", securityHandler, organizationService.UpdateOrganization)
	g.DELETE(organizationCRUD+"/:id", securityHandler, organizationService.DeleteOrganization)
}
