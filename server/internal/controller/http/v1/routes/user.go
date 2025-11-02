package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	userCRUD = "/user"
)

type IUserService interface {
	CreateUser(c *gin.Context)
	GetUser(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	RestoreUser(c *gin.Context)
	InviteUser(c *gin.Context)
	RegisterUser(c *gin.Context)
	InviteUserAcceptHandler(c *gin.Context)
}

func RegisterUserRoutes(g *gin.RouterGroup, userService IUserService, securityHandler gin.HandlerFunc) {
	//g.POST(userCRUD, userService.CreateUser)
	g.POST(userCRUD+"/register", userService.RegisterUser)
	g.GET(userCRUD+"/:id", securityHandler, userService.GetUser)
	g.POST(userCRUD+"/invite", securityHandler, userService.InviteUser)
	g.GET(userCRUD+"/:id/invite/accept", userService.InviteUserAcceptHandler)
	g.PUT(userCRUD+"/:id", securityHandler, userService.UpdateUser)
	g.DELETE(userCRUD+"/:id", securityHandler, userService.DeleteUser)
	g.DELETE(userCRUD+"/:id/restore", securityHandler, userService.RestoreUser)

}
