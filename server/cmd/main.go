package main

import (
	"context"
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
	"log"
	"net/http"
	"os"
	"os/signal"
	_ "semki/docs"
	"semki/internal/adapter/mongo"
	"semki/internal/adapter/qdrant"
	"semki/internal/controller/http/v1/routes"
	"semki/internal/service"
	"semki/internal/utils/config"
	"semki/internal/utils/jwtUtils"
	"semki/pkg/clients"
	google2 "semki/pkg/google"
	"semki/pkg/telemetry"
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
// @description				Type "Bearer" followed by a space and JWT token.
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

	vectorDb, err := qdrant.SetupQdrant(&cfg.Qdrant)
	if err != nil {
		log.Fatalf("Failed to create Qdrant client: %v", err)
	}
	telemetry.Log.Info("Connected to Qdrant")
	defer vectorDb.Close()

	if cfg.EnabledPyroscope {
		pyroscope, err := telemetry.SetupPyroscope(cfg.PyroscopeAddress)
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

	redis := clients.ConnectToRedis(cfg)
	defer redis.Close()

	qdrantRepo := qdrant.New(cfg, vectorDb)

	statusRepo := mongo.NewStatusRepository(db)
	chatRepo := mongo.NewChatRepository(db)
	userRepo := mongo.NewUserRepository(cfg, db)
	orgRepo := mongo.NewOrganizationRepository(db)

	statusService := service.NewStatusService(statusRepo)
	emailService := service.NewEmailService(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		cfg.SMTP.FromName,
	)
	llmService := service.NewLLMService(cfg.OpenAIKey)
	chatService := service.NewChatService(chatRepo, userRepo)
	authService := service.NewAuthService(userRepo)
	embedderService := service.NewEmbedderService(cfg.Embedder.Url)
	qdrantService := service.NewQdrantService(qdrantRepo, userRepo, embedderService)
	organizationService := service.NewOrganizationService(orgRepo, userRepo, qdrantService)
	authMiddleware := jwtUtils.Startup(cfg, authService)
	withAuth := jwtUtils.UseAuth(authMiddleware, cfg, redis)
	logoutHandler := jwtUtils.LogoutHandler(authMiddleware, cfg, redis)
	searchService := service.NewSearchService(qdrantService, llmService, orgRepo, chatRepo, userRepo, telemetry.Log)
	userService := service.NewUserService(qdrantService, userRepo, orgRepo, emailService, authMiddleware, cfg)

	var googleAuthService routes.IGoogleAuthService
	if cfg.Google.Enabled {
		google := google2.InitGoogleOAuth(
			cfg.Protocol+"://"+cfg.Host+":"+cfg.Port+"/api/v1"+routes.GoogleCallback,
			cfg.Google.ClientID,
			cfg.Google.ClientSecret)
		googleAuthService = service.NewGoogleAuthService(qdrantService, userRepo, google, authMiddleware, cfg.FrontendUrl)
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
	corsCfg.AddAllowHeaders(jwtUtils.AuthorizationHeader)
	r.Use(cors.New(corsCfg))

	r.Use(otelgin.Middleware(cfg.Service))
	r.Use(gin.Logger())
	//r.Use(telemetry.LoggerMiddleware(telemetry.Log))
	r.Use(telemetry.TraceIDMiddleware())
	// endregion

	// region ROUTES
	registerSwagger(r)

	apiV1 := r.Group("/api/v1")
	{
		routes.RegisterStatusRoutes(apiV1, statusService)
		routes.RegisterUserRoutes(apiV1, userService, withAuth)
		routes.RegisterOrganizationRoutes(apiV1, organizationService, withAuth)
		routes.RegisterAuthRoutes(apiV1, authService, googleAuthService, authMiddleware, withAuth, logoutHandler)
		routes.RegisterSearchRoutes(apiV1, searchService, withAuth, redis)
		routes.RegisterChatRoutes(apiV1, chatService, withAuth, redis)
		routes.RegisterQdrantRoutes(apiV1, withAuth, qdrantService)
	}
	r.NoRoute(jwtUtils.NoRoute)
	// endregion

	// region IGNITION
	telemetry.Log.Info(fmt.Sprintf("IGNITION at http://localhost:%s/swagger/index.html", cfg.Port))
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  5 * time.Minute,
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

func registerSwagger(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
		ginSwagger.DocExpansion("list"),
		ginSwagger.DeepLinking(true),
	))
}
