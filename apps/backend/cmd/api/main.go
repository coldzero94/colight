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

	var crawlingService *service.CrawlingService
	if aiProvider != nil {
		crawlingService = service.NewCrawlingService(aiProvider)
	}

	companyDataService := service.NewCompanyDataService()

	var companyAnalysisService *service.CompanyAnalysisService
	if aiProvider != nil {
		companyAnalysisService = service.NewCompanyAnalysisService(db, aiProvider, companyDataService)
	}

	var matchingService *service.MatchingService
	if aiProvider != nil {
		matchingService = service.NewMatchingService(db, aiProvider.Light())
	}

	var questionService *service.QuestionService
	if aiProvider != nil {
		questionService = service.NewQuestionService(db, aiProvider.Heavy())
	}

	var coachingService *service.CoachingService
	if aiProvider != nil {
		coachingService = service.NewCoachingService(db, aiProvider.HeavyStreaming())
	}

	var reviewService *service.ReviewService
	if aiProvider != nil {
		reviewService = service.NewReviewService(db, aiProvider.Heavy())
	}

	editorService := service.NewEditorService(db)

	// Controllers
	authCtrl := controller.NewAuthController(authService, cfg)
	adminCtrl := controller.NewAdminController(db)
	experienceCtrl := controller.NewExperienceController(experienceService)

	var weaponTaggingCtrl *controller.WeaponTaggingController
	if weaponTaggingService != nil {
		weaponTaggingCtrl = controller.NewWeaponTaggingController(weaponTaggingService)
	}

	var crawlingCtrl *controller.CrawlingController
	if crawlingService != nil {
		crawlingCtrl = controller.NewCrawlingController(crawlingService)
	}

	companyDataCtrl := controller.NewCompanyDataController(companyDataService)

	var companyAnalysisCtrl *controller.CompanyAnalysisController
	if companyAnalysisService != nil {
		companyAnalysisCtrl = controller.NewCompanyAnalysisController(companyAnalysisService)
	}

	var matchingCtrl *controller.MatchingController
	if matchingService != nil && companyAnalysisService != nil {
		matchingCtrl = controller.NewMatchingController(matchingService, companyAnalysisService)
	}

	var questionCtrl *controller.QuestionController
	if questionService != nil {
		questionCtrl = controller.NewQuestionController(questionService)
	}

	var coachingCtrl *controller.CoachingController
	if coachingService != nil {
		coachingCtrl = controller.NewCoachingController(coachingService)
	}

	var reviewCtrl *controller.ReviewController
	if reviewService != nil {
		reviewCtrl = controller.NewReviewController(reviewService)
	}

	editorCtrl := controller.NewEditorController(editorService)

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

		// Job posting crawling (if AI provider is available)
		if crawlingCtrl != nil {
			protected.POST("/crawl", crawlingCtrl.ParseJobPosting)
		}

		// Company data (DART + News crawling)
		protected.GET("/company-data", companyDataCtrl.GetCompanyData)

		// Company analysis (AI-powered talent profile analysis)
		if companyAnalysisCtrl != nil {
			protected.POST("/analyze-company", companyAnalysisCtrl.AnalyzeCompany)
		}

		// Experience matching (AI-powered experience-company matching)
		if matchingCtrl != nil {
			protected.POST("/match", matchingCtrl.MatchExperiences)
		}

		// Applications list (for coaching page company select)
		if questionCtrl != nil {
			protected.GET("/applications", questionCtrl.GetApplications)
		}

		// Question analysis (AI-powered question intent analysis)
		if questionCtrl != nil {
			protected.POST("/coaching/question-analysis", questionCtrl.PostQuestionAnalysis)
			protected.POST("/coaching/recommend-experiences", questionCtrl.PostRecommendExperiences)
		}

		// Draft coaching (AI-powered draft generation)
		if coachingCtrl != nil {
			protected.POST("/coaching/draft", coachingCtrl.PostDraft)
			protected.GET("/coaching/sessions", coachingCtrl.GetSessions)
		}

		// Review coaching (AI-powered cover letter review)
		if reviewCtrl != nil {
			protected.POST("/coaching/review", reviewCtrl.PostReview)
		}

		// Cover letter editor (CRUD + versioning)
		protected.GET("/coaching/cover-letters/:id", editorCtrl.GetCoverLetter)
		protected.PATCH("/coaching/cover-letters/:id", editorCtrl.PatchCoverLetter)
		protected.POST("/coaching/cover-letters/:id/versions", editorCtrl.PostVersion)
		protected.GET("/coaching/cover-letters/:id/versions", editorCtrl.GetVersions)
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
