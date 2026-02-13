package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/coby/colight/apps/backend/internal/controller"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
	"github.com/coby/colight/apps/backend/internal/infrastructure/database"
	"github.com/coby/colight/apps/backend/internal/infrastructure/middleware"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.NewClient(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("connected to database")

	// AI Provider
	ctx := context.Background()
	aiProvider, err := ai.NewAIProvider(ctx, cfg)
	if err != nil {
		slog.Warn("failed to initialize AI provider", "error", err)
		// AI provider is optional - services will handle nil gracefully
		aiProvider = nil
	}
	if aiProvider != nil {
		defer aiProvider.Close()
	}

	// Services
	tokenService := service.NewTokenService(cfg.JWTSecret, cfg.JWTAccessTokenTTL, cfg.JWTRefreshTokenTTL)
	authService := service.NewAuthService(cfg, db, tokenService)
	experienceService := service.NewExperienceService(db)

	var weaponTaggingService *service.WeaponTaggingService
	if aiProvider != nil {
		weaponTaggingService = service.NewWeaponTaggingService(db, aiProvider.Light())
	}

	// Controllers
	authCtrl := controller.NewAuthController(authService, cfg)
	adminCtrl := controller.NewAdminController(db)
	experienceCtrl := controller.NewExperienceController(experienceService)

	var weaponTaggingCtrl *controller.WeaponTaggingController
	if weaponTaggingService != nil {
		weaponTaggingCtrl = controller.NewWeaponTaggingController(weaponTaggingService)
	}

	// Router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware(cfg.FrontendURL))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public auth routes
	auth := r.Group("/v1/auth")
	{
		auth.GET("/naver/login", authCtrl.NaverLogin)
		auth.GET("/naver/callback", authCtrl.NaverCallback)
		auth.POST("/signup", authCtrl.Signup)
		auth.POST("/login", authCtrl.Login)
		auth.POST("/refresh", authCtrl.Refresh)
	}

	// Protected routes (require authentication)
	protected := r.Group("/v1")
	protected.Use(middleware.AuthMiddleware(tokenService))
	{
		protected.GET("/auth/me", authCtrl.Me)
		protected.POST("/auth/logout", authCtrl.Logout)

		// Experience CRUD
		protected.POST("/experiences", experienceCtrl.Create)
		protected.GET("/experiences", experienceCtrl.List)
		protected.GET("/experiences/:id", experienceCtrl.Get)
		protected.PATCH("/experiences/:id", experienceCtrl.Update)
		protected.DELETE("/experiences/:id", experienceCtrl.Delete)

		// Weapon tagging (if AI provider is available)
		if weaponTaggingCtrl != nil {
			protected.POST("/experiences/:id/tag", weaponTaggingCtrl.Tag)
		}
	}

	// Admin routes (require authentication + admin role)
	admin := r.Group("/v1/admin")
	admin.Use(middleware.AuthMiddleware(tokenService))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/users", adminCtrl.ListUsers)
		admin.GET("/users/:id", adminCtrl.GetUser)
		admin.PUT("/users/:id/role", adminCtrl.UpdateUserRole)
		admin.GET("/stats", adminCtrl.GetStats)
		admin.GET("/prompts", adminCtrl.ListPrompts)
		admin.PUT("/prompts/:id", adminCtrl.UpdatePrompt)
	}

	slog.Info("starting colight api server", "port", cfg.APIPort)
	if err := r.Run(":" + cfg.APIPort); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
