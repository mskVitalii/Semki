package main

import (
	"context"
	_ "dwt/docs"
	"dwt/internal/adapter/mongo"
	"dwt/internal/controller/http/v1/routes"
	"dwt/internal/service"
	"dwt/internal/utils/config"
	"dwt/internal/utils/jwt"
	"dwt/pkg/clients"
	google2 "dwt/pkg/google"
	"dwt/pkg/telemetry"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// @title						Semki
// @version					1.0
// @description				Semantic Contacts
// @contact.name				Vitalii Popov
// @contact.url				https://www.linkedin.com/in/mskVitalii/
// @contact.email				msk.vitaly@gmail.com
// @BasePath					/
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	cfg := config.GetConfig("")
	startup(cfg)
}

func startup(cfg *config.Config) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// region DATABASES & SERVICES
	telemetry.SetupLogger(cfg)
	defer telemetry.Log.Sync()

	otelShutdown, err := telemetry.SetupOTelSDK(ctx, cfg)
	if err != nil {
		telemetry.Log.Fatal("[startup] Failed to setup otel SDK", zap.Error(err))
	}
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	if cfg.EnabledPyroscope {
		pyroscope, err := telemetry.SetupPyroscope(cfg.PyroscopeServerAddress)
		if err != nil {
			telemetry.Log.Fatal("[SetupPyroscope] Error setting up pyroscope", zap.Error(err))
			return
		}
		defer pyroscope.Stop()
	}

	db, err := mongo.SetupMongo(&cfg.Mongo)
	if err != nil {
		telemetry.Log.Fatal("failed to connect MongoDB", zap.Error(err))
	}
	if mongo.IsPlacesCollectionExist(db) == false {
		// TODO: JSON mock for organization
		// TODO: Add route to upload mock
		telemetry.Log.Info("successfully loaded datasets")
	} else {
		telemetry.Log.Info("datasets are in database")
	}

	mongoRepo := mongo.New(cfg, db)
	statusService := service.NewStatusService()
	userService := service.NewUserService(mongoRepo)
	authService := service.NewAuthService(mongoRepo)
	auth := jwt.Startup(cfg, authService)
	var googleAuthService routes.IGoogleAuthService
	if cfg.Google.Enabled {
		google := google2.InitGoogleOAuth(
			cfg.Protocol+"://"+cfg.Host+":"+cfg.Port+"/api/v1"+routes.GoogleCallback,
			cfg.Google.ClientID,
			cfg.Google.ClientSecret)
		googleAuthService = service.NewGoogleAuthService(mongoRepo, google, auth, cfg.FrontendUrl)
	}
	// endregion

	// region GIN
	if cfg.IsDebug == false {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	if cfg.EnabledSentry {
		clients.SetupSentry(cfg.SentryDSN, cfg.SentryEnableTracing, cfg.SentryTracesSampleRate)
		r.Use(sentrygin.New(sentrygin.Options{}))
	}

	telemetry.SetupPrometheus(r, cfg.GrafanaSlowRequest)

	// Error recovery middleware with Sentry
	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Capture the panic
				sentry.CurrentHub().Recover(err)
				sentry.Flush(time.Second * 5)
				// Return a 500 Internal Server Error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	})

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = []string{
		cfg.FrontendUrl,
		"http://prometheus:9090",
		"https://*.semki.local",
		"https://api.semki.local",
		"http://localhost:8080"}
	corsCfg.AllowCredentials = true
	corsCfg.AddExposeHeaders(telemetry.TraceHeader)
	corsCfg.AddAllowHeaders(jwt.AuthorizationHeader)
	r.Use(cors.New(corsCfg))

	r.Use(otelgin.Middleware(cfg.Service))
	r.Use(gin.Logger())
	//r.Use(telemetry.LoggerMiddleware(telemetry.Log))
	r.Use(telemetry.TraceIDMiddleware())
	// endregion

	// region ROUTES
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := r.Group("/api/v1")
	{
		routes.RegisterStatusRoutes(apiV1, statusService)
		routes.RegisterUserRoutes(apiV1, userService, auth)
		routes.RegisterAuthRoutes(apiV1, authService, googleAuthService, auth)
	}
	r.NoRoute(auth.MiddlewareFunc(), jwt.NoRoute)
	// endregion

	// region IGNITION
	telemetry.Log.Info(fmt.Sprintf("IGNITION at http://localhost:%s/swagger/index.html", cfg.Port))
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		sentry.CaptureException(err)
		sentry.Flush(time.Second * 5)
		telemetry.Log.Fatal("Ignition error: ", zap.Error(err))
	case <-ctx.Done():
		stop()
	}

	err = srv.Shutdown(context.Background())
	return
	// endregion
}
