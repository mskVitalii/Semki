package routes

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/model"
	"semki/pkg/lib"
)

const (
	login          = "/login"
	logout         = "/logout"
	refreshToken   = "/refresh_token"
	googleLogin    = "/google/login"
	GoogleCallback = "/google/callback"
	claims         = "/claims"
)

type IAuthService interface {
	LoginHandler(c *gin.Context)
	LogoutHandler(c *gin.Context)
	RefreshTokenHandler(c *gin.Context)
	ClaimsHandler(c *gin.Context)
	Authenticate(request dto.LoginRequest) (*model.User, error)
}

type IGoogleAuthService interface {
	GoogleLoginHandler(c *gin.Context)
	GoogleAuthCallback(c *gin.Context)
}

func RegisterAuthRoutes(g *gin.RouterGroup,
	authService IAuthService,
	googleService IGoogleAuthService,
	authMiddleware *jwt.GinJWTMiddleware,
	withAuth gin.HandlerFunc,
	logoutHandler gin.HandlerFunc) {

	g.POST(login, authMiddleware.LoginHandler)
	g.POST(logout, withAuth, logoutHandler)
	g.POST(refreshToken, authMiddleware.RefreshHandler)
	g.GET(googleLogin, googleService.GoogleLoginHandler)
	g.OPTIONS(googleLogin, lib.Preflight)
	g.GET(GoogleCallback, googleService.GoogleAuthCallback)
	g.GET(claims, withAuth, authService.ClaimsHandler)
}
