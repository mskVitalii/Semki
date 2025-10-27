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
	"semki/internal/utils/jwtUtils"
	"semki/pkg/google"
	"semki/pkg/telemetry"
)

// authService - dependent services
type googleAuthService struct {
	repo        mongo.IRepository
	google      google.Google
	jwtAuth     *jwt.GinJWTMiddleware
	frontendUrl string
}

func NewGoogleAuthService(
	repo mongo.IRepository,
	google google.Google,
	jwtAuth *jwt.GinJWTMiddleware,
	frontendUrl string,
) routes.IGoogleAuthService {

	return &googleAuthService{repo, google, jwtAuth, frontendUrl}
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
	userFromDb, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20db")
		return
	}

	if userFromDb == nil {
		userFromDb = dto.NewUserFromGoogleProvider(user)
		if err := s.repo.CreateUser(ctx, userFromDb); err != nil {
			c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20create-user")
			return
		}
	} else if model.ProviderInUserProviders(model.UserProviders.Google, userFromDb.Providers) == false {
		userFromDb.Providers = append(userFromDb.Providers, model.UserProviders.Google)
		if err := s.repo.UpdateUser(ctx, userFromDb.ID, *userFromDb); err != nil {
			c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20update%20provider")
			return
		}
	}

	if userFromDb.Status == model.UserStatuses.DELETED {
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=user%20deleted")
		return
	}

	// Token
	claims, err := jwtUtils.UserToPayload(userFromDb)
	jwtToken, err := s.jwtAuth.TokenGenerator(claims)
	if err != nil {
		telemetry.Log.Error(err.Error())
		c.Redirect(http.StatusFound, s.frontendUrl+"/login?error=internal%20error%20token")
		return
	}

	c.Redirect(http.StatusFound, s.frontendUrl+"/login?accessToken="+jwtToken.AccessToken+"&refresh="+jwtToken.RefreshToken)
}
