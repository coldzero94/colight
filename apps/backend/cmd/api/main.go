package main

import (
	"log/slog"
	"os"

	"github.com/coby/colight/apps/backend/internal/controller"
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

	// Services
	tokenService := service.NewTokenService(cfg.JWTSecret, cfg.JWTAccessTokenTTL, cfg.JWTRefreshTokenTTL)
	authService := service.NewAuthService(cfg, db, tokenService)

	// Controllers
	authCtrl := controller.NewAuthController(authService, cfg)
	adminCtrl := controller.NewAdminController(db)

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
