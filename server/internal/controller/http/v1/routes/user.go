package routes

import (
	"github.com/gin-gonic/gin"
	"semki/pkg/lib"
)

const (
	userCRUD      = "/user"
	resetPassword = "/reset_password"
)

type IUserService interface {
	CreateUser(c *gin.Context)
	GetUser(c *gin.Context)
	UpdateUser(c *gin.Context)
	PatchUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	RestoreUser(c *gin.Context)
	InviteUser(c *gin.Context)
	RegisterUser(c *gin.Context)
	InviteUserAcceptHandler(c *gin.Context)
	VerifyUserEmailHandler(c *gin.Context)
	SetPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	ConfirmResetPasswordHandler(c *gin.Context)
}

func RegisterUserRoutes(g *gin.RouterGroup, userService IUserService, securityHandler gin.HandlerFunc) {
	//g.POST(userCRUD, userService.CreateUser)
	g.POST(userCRUD+"/register", userService.RegisterUser)
	g.POST(userCRUD+"/invite", securityHandler, userService.InviteUser)
	g.GET(userCRUD+"/:id/invite/accept", userService.InviteUserAcceptHandler)
	g.GET(userCRUD+"/:id/verify/accept", userService.VerifyUserEmailHandler)
	g.OPTIONS(userCRUD+"/set_password", lib.Preflight)
	g.POST(userCRUD+"/set_password", securityHandler, userService.SetPassword)
	g.PUT(userCRUD+"/:id", securityHandler, userService.UpdateUser)
	g.PATCH(userCRUD+"/:id", securityHandler, userService.PatchUser)
	g.GET(userCRUD+"/:id", securityHandler, userService.GetUser)
	g.DELETE(userCRUD+"/:id", securityHandler, userService.DeleteUser)
	g.POST(userCRUD+"/:id/restore", securityHandler, userService.RestoreUser)
	g.POST(userCRUD+resetPassword, userService.ResetPassword)
	g.OPTIONS(userCRUD+resetPassword+"/confirm", lib.Preflight)
	g.GET(userCRUD+resetPassword+"/confirm", userService.ConfirmResetPasswordHandler)
}
