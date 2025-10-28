package jwtUtils

import (
	"errors"
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"semki/internal/controller/http/v1/dto"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/model"
	"semki/internal/utils/config"
	"semki/pkg/lib"
	"semki/pkg/telemetry"
	"strings"
	"time"
)

const (
	IdentityKey         = "_id"
	AuthorizationHeader = "Authorization"
	AccessTokenTimeout  = time.Minute * 15
	RefreshTokenTimeout = time.Hour * 24 * 7
)

func Startup(cfg *config.Config, service routes.IAuthService) *jwt.GinJWTMiddleware {
	middleware := &jwt.GinJWTMiddleware{
		Realm:           cfg.Service,
		Key:             []byte(cfg.SecretKeyJWT),
		Timeout:         AccessTokenTimeout,
		MaxRefresh:      RefreshTokenTimeout,
		IdentityKey:     IdentityKey,
		IdentityHandler: identity,
		Authenticator:   authenticator(service),
		Authorizer:      authorization,
		Unauthorized:    unauthorized,
		PayloadFunc:     payloadFunc,
		TokenLookup:     fmt.Sprintf("header:%s", AuthorizationHeader),
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	}

	middleware.EnableRedisStore(
		jwt.WithRedisAddr(fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)),
		jwt.WithRedisAuth(cfg.Redis.Password, 0),
		jwt.WithRedisKeyPrefix(cfg.Service+":jwt:"),
		jwt.WithRedisCache(64*1024*1024, 30*time.Second))

	authMiddleware, err := jwt.New(middleware)
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	if errInit := authMiddleware.MiddlewareInit(); errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	return authMiddleware
}

func authenticator(service routes.IAuthService) func(*gin.Context) (interface{}, error) {
	return func(c *gin.Context) (interface{}, error) {
		var loginVals dto.LoginRequest
		if err := c.ShouldBind(&loginVals); err != nil {
			return "", errors.New("missing email or password")
		}
		if lib.IsValidEmail(loginVals.Email) == false {
			return "", errors.New("invalid email")
		}
		user, err := service.Authenticate(loginVals)
		if err != nil {
			telemetry.Log.Info(err.Error())
			return nil, jwt.ErrFailedAuthentication
		}

		//c.Set(IdentityKey, user)
		return user, nil
	}
}

func NoRoute(c *gin.Context) {
	//claims := jwt.ExtractClaims(c)
	//telemetry.Log.Info(fmt.Sprintf("NoRoute claims: %v", claims), telemetry.TraceForZapLog(c.Request.Context()))
	c.JSON(http.StatusNotFound, gin.H{"message": "Not found"})
}

// region Payload

type UserClaims struct {
	ID               primitive.ObjectID     `json:"_id"`
	OrganizationId   primitive.ObjectID     `json:"organizationId"`
	OrganizationRole model.OrganizationRole `json:"organizationRole"`
}

func UserToPayload(data interface{}) (*UserClaims, error) {
	// login flow: authenticator -> payloadFunc
	user, ok := data.(*model.User)
	if !ok {
		// refresh flow: claims in Redis -> payloadFunc
		b, err := bson.Marshal(data)
		if err != nil {
			telemetry.Log.Error("UserToPayload Error bson.Marshal")
			return nil, err
		}
		var u model.User
		if err := bson.Unmarshal(b, &u); err != nil {
			telemetry.Log.Error("UserToPayload Error bson.Unmarshal")
			return nil, err
		}
		user = &u
		//return &UserClaims{}, errors.New("invalid user data")
	}
	return &UserClaims{
		ID:               user.ID,
		OrganizationId:   user.OrganizationID,
		OrganizationRole: user.OrganizationRole,
	}, nil
}

func payloadFunc(data interface{}) gojwt.MapClaims {
	if v, err := UserToPayload(data); err == nil {
		return gojwt.MapClaims{
			IdentityKey:        v.ID.Hex(),
			"organizationId":   v.OrganizationId.Hex(),
			"organizationRole": string(v.OrganizationRole),
		}
	}
	return gojwt.MapClaims{}
}

// endregion

// region Identity

func identity(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)

	idStr, ok := claims[IdentityKey].(string)
	if !ok {
		return nil
	}

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil
	}

	orgIdStr, ok := claims["organizationId"].(string)
	if !ok {
		return nil
	}

	orgId, err := primitive.ObjectIDFromHex(orgIdStr)
	if err != nil {
		return nil
	}

	orgRole, _ := claims["organizationRole"].(string)

	return &UserClaims{
		ID:               id,
		OrganizationId:   orgId,
		OrganizationRole: model.OrganizationRole(orgRole),
	}
}

// endregion

// region Authorization

func authorization(_ *gin.Context, _ any) bool {
	// TODO: check organization by gin.Context & data
	return true
	//return data != nil
}

func unauthorized(c *gin.Context, code int, message string) {
	if v, exists := c.Get("auth_error"); exists && v == "blacklisted" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is blacklisted"})
		return
	}
	c.JSON(code, dto.UnauthorizedResponse{Message: message})
}

// endregion

// region Blacklist

func UseAuth(auth *jwt.GinJWTMiddleware, cfg *config.Config, rdb *redis.Client) gin.HandlerFunc {
	prefix := cfg.Service + ":jwt:bl:"
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 {
				token := parts[1]
				key := prefix + token
				if exists, err := rdb.Exists(c.Request.Context(), key).Result(); err == nil && exists > 0 {
					c.Set("auth_error", "blacklisted")
					unauthorized(c, http.StatusUnauthorized, "token is blacklisted")
					c.Abort()
					return
				}
			}
		}

		auth.MiddlewareFunc()(c)
		if c.IsAborted() {
			return
		}
	}
}

func LogoutHandler(auth *jwt.GinJWTMiddleware, cfg *config.Config, rdb *redis.Client) gin.HandlerFunc {
	prefix := cfg.Service + ":jwt:"
	prefixBl := prefix + "bl:"

	return func(c *gin.Context) {
		var req dto.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token required"})
			return
		}

		refreshKey := prefix + req.RefreshToken
		rdb.Del(c.Request.Context(), refreshKey)

		authHeader := c.GetHeader(AuthorizationHeader)
		parts := strings.SplitN(authHeader, " ", 2)
		token := parts[1]
		key := prefixBl + token
		rdb.Set(c.Request.Context(), key, "1", AccessTokenTimeout)

		auth.LogoutHandler(c)
	}
}

// endregion
