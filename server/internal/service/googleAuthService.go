package service

import (
	"encoding/json"
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"net/http"
	"semki/internal/adapter/mongo"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/pkg/google"
	"semki/pkg/telemetry"
)

// authService - dependent services
type googleAuthService struct {
	qdrantService IQdrantService
	repo          mongo.IUserRepository
	google        google.Google
	jwtAuth       *jwt.GinJWTMiddleware
	frontendUrl   string
}

func NewGoogleAuthService(
	qdrantService IQdrantService,
	repo mongo.IUserRepository,
	google google.Google,
	jwtAuth *jwt.GinJWTMiddleware,
	frontendUrl string,
) routes.IGoogleAuthService {

	return &googleAuthService{qdrantService, repo, google, jwtAuth, frontendUrl}
}

// GoogleLoginHandler godoc
//
//	@Summary	Used by Google Auth provider
//	@Tags		auth
//	@Router		/api/v1/google/login [get]
func (s *googleAuthService) GoogleLoginHandler(c *gin.Context) {
	url := s.google.OAuthConfig.AuthCodeURL("state-string", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleAuthCallback godoc
//
//	@Summary	Used by Google Auth provider
//	@Tags		auth
//	@Router		/api/v1/google/callback [get]
func (s *googleAuthService) GoogleAuthCallback(c *gin.Context) {
	state := c.Query("state")
	if state != "state-string" {
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error")
		return
	}

	ctx := c.Request.Context()
	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=google-error-code")
		return
	}
	token, err := s.google.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		telemetry.Log.Error("Exchange error: " + err.Error())
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=google%20error%20token")
		return
	}

	client := s.google.OAuthConfig.Client(ctx, token)
	userInfo, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		telemetry.Log.Error("Userinfo error" + err.Error())
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=google%20error%20userinfo")
		return
	}
	defer userInfo.Body.Close()

	var user dto.CreateUserByGoogleProvider
	if err = json.NewDecoder(userInfo.Body).Decode(&user); err != nil {
		telemetry.Log.Error("Userinfo error" + err.Error())
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20body")
		return
	}

	// DB
	userFromDB, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20db")
		return
	}

	if userFromDB == nil {
		userFromDB = dto.NewUserFromGoogleProvider(user)
		if userFromDB.OrganizationRole == "" {
			// Invited users has its Role. Probably fix it later
			userFromDB.OrganizationRole = model.OrganizationRoles.OWNER
		}
		if err := s.repo.CreateUser(ctx, userFromDB); err != nil {
			c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20create-user")
			return
		}
	} else if model.ProviderInUserProviders(model.UserProviders.Google, userFromDB.Providers) == false {
		if userFromDB.Status == model.UserStatuses.DELETED {
			c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=user%20deleted")
			return
		}
		userFromDB.Providers = append(userFromDB.Providers, model.UserProviders.Google)
		if err := s.repo.UpdateUser(ctx, userFromDB.ID, *userFromDB); err != nil {
			c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20update%20provider")
			return
		}
	}

	if err := s.qdrantService.IndexUser(ctx, userFromDB); err != nil {
		telemetry.Log.Error("Failed to index created user from Google in Qdrant: " + err.Error())
	}

	// Token
	jwtToken, err := s.jwtAuth.TokenGenerator(userFromDB)
	if err != nil {
		telemetry.Log.Error(err.Error())
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20token")
		return
	}

	c.Redirect(http.StatusFound, s.frontendUrl+"/login?accessToken="+jwtToken.AccessToken+"&refreshToken="+jwtToken.RefreshToken)
}
