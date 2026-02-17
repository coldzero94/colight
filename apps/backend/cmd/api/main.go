package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/coby/colight/apps/backend/internal/controller"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
	"github.com/coby/colight/apps/backend/internal/infrastructure/database"
	"github.com/coby/colight/apps/backend/internal/infrastructure/logger"
	"github.com/coby/colight/apps/backend/internal/infrastructure/middleware"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (optional - ignore if not found)
	_ = godotenv.Load("../../.env")

	cfg := config.Load()

	// Initialize structured logger and set as default
	appLogger := logger.New(cfg.AppEnv, cfg.LogLevel)
	slog.SetDefault(appLogger)

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
		weaponTaggingService = service.NewWeaponTaggingService(db, aiProvider)
	}

	var starGenerationService *service.StarGenerationService
	if aiProvider != nil {
		starGenerationService = service.NewStarGenerationService(db, aiProvider)
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
		matchingService = service.NewMatchingService(db, aiProvider)
	}

	var questionService *service.QuestionService
	if aiProvider != nil {
		questionService = service.NewQuestionService(db, aiProvider)
	}

	var coachingService *service.CoachingService
	if aiProvider != nil {
		coachingService = service.NewCoachingService(db, aiProvider)
	}

	var reviewService *service.ReviewService
	if aiProvider != nil {
		reviewService = service.NewReviewService(db, aiProvider)
	}

	var charCoachingService *service.CharCoachingService
	if aiProvider != nil {
		charCoachingService = service.NewCharCoachingService(db, aiProvider)
	}

	var interviewService *service.InterviewService
	if aiProvider != nil {
		interviewService = service.NewInterviewService(aiProvider, db, weaponTaggingService)
	}

	editorService := service.NewEditorService(db)
	usageService := service.NewUsageService(db)
	feedbackService := service.NewFeedbackService(db)
	applicationService := service.NewApplicationService(db)

	// Controllers
	authCtrl := controller.NewAuthController(authService, cfg)
	adminCtrl := controller.NewAdminController(db)
	experienceCtrl := controller.NewExperienceController(experienceService)
	usageCtrl := controller.NewUsageController(usageService)

	var weaponTaggingCtrl *controller.WeaponTaggingController
	if weaponTaggingService != nil {
		weaponTaggingCtrl = controller.NewWeaponTaggingController(weaponTaggingService)
	}

	var starGenCtrl *controller.StarGenerationController
	if starGenerationService != nil {
		starGenCtrl = controller.NewStarGenerationController(starGenerationService)
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

	var charCoachingCtrl *controller.CharCoachingController
	if charCoachingService != nil {
		charCoachingCtrl = controller.NewCharCoachingController(charCoachingService)
	}

	var interviewCtrl *controller.InterviewController
	if interviewService != nil {
		interviewCtrl = controller.NewInterviewController(interviewService)
	}

	editorCtrl := controller.NewEditorController(editorService)
	feedbackCtrl := controller.NewFeedbackController(feedbackService)
	applicationCtrl := controller.NewApplicationController(applicationService)

	// Router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(appLogger))
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

		// Usage tracking
		protected.GET("/usage", usageCtrl.GetUsage)

		// Experience CRUD (create has usage limit)
		protected.POST("/experiences", controller.UsageLimitMiddleware(usageService, "experience"), experienceCtrl.Create)
		protected.GET("/experiences", experienceCtrl.List)
		protected.GET("/experiences/:id", experienceCtrl.Get)
		protected.PATCH("/experiences/:id", experienceCtrl.Update)
		protected.DELETE("/experiences/:id", experienceCtrl.Delete)

		// Weapon tagging (if AI provider is available)
		if weaponTaggingCtrl != nil {
			protected.POST("/experiences/:id/tag", weaponTaggingCtrl.Tag)
		}

		// AI STAR generation from free-form content
		if starGenCtrl != nil {
			protected.POST("/experiences/generate-star", starGenCtrl.GenerateSTAR)
		}

		// Job posting crawling (if AI provider is available)
		if crawlingCtrl != nil {
			protected.POST("/crawl", crawlingCtrl.ParseJobPosting)
		}

		// Company data (DART + News crawling)
		protected.GET("/company-data", companyDataCtrl.GetCompanyData)

		// Company analysis (AI-powered talent profile analysis)
		if companyAnalysisCtrl != nil {
			protected.POST("/analyze-company", controller.UsageLimitMiddleware(usageService, "analysis"), companyAnalysisCtrl.AnalyzeCompany)
		}

		// Experience matching (AI-powered experience-company matching)
		if matchingCtrl != nil {
			protected.POST("/match", matchingCtrl.MatchExperiences)
		}

		// Applications (dashboard + coaching page)
		protected.GET("/applications", applicationCtrl.ListApplications)
		protected.PATCH("/applications/:id/status", applicationCtrl.UpdateStatus)
		protected.GET("/applications/stats", applicationCtrl.GetStats)

		// Question analysis (AI-powered question intent analysis)
		if questionCtrl != nil {
			protected.POST("/coaching/question-analysis", controller.UsageLimitMiddleware(usageService, "question_analysis"), questionCtrl.PostQuestionAnalysis)
			protected.POST("/coaching/recommend-experiences", questionCtrl.PostRecommendExperiences)
		}

		// Draft coaching (AI-powered draft generation)
		if coachingCtrl != nil {
			protected.POST("/coaching/draft", controller.UsageLimitMiddleware(usageService, "draft"), coachingCtrl.PostDraft)
			protected.GET("/coaching/sessions", coachingCtrl.GetSessions)
		}

		// Review coaching (AI-powered cover letter review)
		if reviewCtrl != nil {
			protected.POST("/coaching/review", controller.UsageLimitMiddleware(usageService, "review"), reviewCtrl.PostReview)
		}

		// Character count coaching (AI-powered trim/expand suggestions)
		if charCoachingCtrl != nil {
			protected.POST("/coaching/char-count", controller.UsageLimitMiddleware(usageService, "char_coaching"), charCoachingCtrl.PostCharCoaching)
		}

		// AI Interview
		if interviewCtrl != nil {
			protected.POST("/interview/question", interviewCtrl.PostQuestion)
			protected.POST("/interview/extract", interviewCtrl.PostExtractSTAR)
			protected.POST("/interview/save", controller.UsageLimitMiddleware(usageService, "experience"), interviewCtrl.PostSaveExperience)
		}

		// Feedback
		protected.POST("/feedback", feedbackCtrl.SubmitFeedback)

		// Cover letter editor (CRUD + versioning)
		protected.GET("/coaching/cover-letters/:id", editorCtrl.GetCoverLetter)
		protected.PATCH("/coaching/cover-letters/:id", editorCtrl.PatchCoverLetter)
		protected.POST("/coaching/cover-letters/:id/versions", editorCtrl.PostVersion)
		protected.GET("/coaching/cover-letters/:id/versions", editorCtrl.GetVersions)
	}

	// Admin routes (require authentication + role-based access)
	admin := r.Group("/v1/admin")
	admin.Use(middleware.AuthMiddleware(tokenService))
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/users", adminCtrl.ListUsers)
		admin.GET("/users/:id", adminCtrl.GetUser)
		admin.PUT("/users/:id/role", adminCtrl.UpdateUserRole)
		admin.GET("/stats", adminCtrl.GetStats)
		admin.GET("/prompts", adminCtrl.ListPrompts)
		admin.PUT("/prompts/:id", adminCtrl.UpdatePrompt)
		admin.GET("/usage/summary", adminCtrl.GetUsageSummary)
		admin.GET("/usage/daily", adminCtrl.GetUsageDaily)
		admin.POST("/users/:id/suspend", adminCtrl.SuspendUser)
		admin.DELETE("/users/:id/suspend", adminCtrl.UnsuspendUser)
		admin.GET("/users/:id/detail", adminCtrl.GetUserDetail)
		admin.GET("/audit-logs", adminCtrl.ListAuditLogs)
		admin.GET("/health", adminCtrl.HealthCheck)
		admin.GET("/feedbacks", adminCtrl.ListFeedbacks)
		admin.PUT("/feedbacks/:id", adminCtrl.UpdateFeedbackStatus)
		admin.GET("/usage/costs", adminCtrl.GetUsageCosts)
		admin.GET("/usage/top-users", adminCtrl.GetUsageTopUsers)
		admin.POST("/users/:id/force-logout", adminCtrl.ForceLogout)
		admin.PUT("/users/:id/plan", adminCtrl.UpdateUserPlan)
		admin.POST("/users/:id/export", adminCtrl.ExportUserData)
	}

	// Super-admin only routes
	superAdmin := r.Group("/v1/admin")
	superAdmin.Use(middleware.AuthMiddleware(tokenService))
	superAdmin.Use(middleware.RequireRole("super_admin"))
	{
		superAdmin.GET("/configs", adminCtrl.ListConfigs)
		superAdmin.PUT("/configs/:key", adminCtrl.UpdateConfig)
		superAdmin.POST("/users/:id/delete-request", adminCtrl.CreateDeletionRequest)
		superAdmin.DELETE("/users/:id/delete-request", adminCtrl.CancelDeletionRequest)
		superAdmin.GET("/deletion-queue", adminCtrl.ListDeletionQueue)
	}

	slog.Info("starting colight api server", "port", cfg.APIPort)
	if err := r.Run(":" + cfg.APIPort); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
