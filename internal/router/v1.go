package router

import (
	"github.com/campus-project-hub/api/internal/handlers"
	"github.com/campus-project-hub/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupV1Routes sets up all API v1 routes
func SetupV1Routes(r *gin.Engine) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	userHandler := handlers.NewUserHandler()
	projectHandler := handlers.NewProjectHandler()
	articleHandler := handlers.NewArticleHandler()
	commentHandler := handlers.NewCommentHandler()
	transactionHandler := handlers.NewTransactionHandler()
	categoryHandler := handlers.NewCategoryHandler()
	uploadHandler := handlers.NewUploadHandler()
	gamificationHandler := handlers.NewGamificationHandler()

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "Campus Project Hub API"})
		})

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/google", authHandler.GoogleAuth)
			auth.GET("/google/callback", authHandler.GoogleCallback)
			auth.GET("/github", authHandler.GitHubAuth)
			auth.GET("/github/callback", authHandler.GitHubCallback)

			// Protected auth routes
			auth.Use(middleware.AuthMiddleware())
			auth.GET("/me", authHandler.GetMe)
			auth.POST("/logout", authHandler.Logout)
		}

		// User routes
		users := api.Group("/users")
		{
			users.GET("/leaderboard", userHandler.Leaderboard)
			users.GET("/:id", userHandler.Get)

			// Protected user routes
			users.Use(middleware.AuthMiddleware())
			users.PUT("/:id", userHandler.Update)

			// Admin only
			users.GET("", middleware.RequireAdmin(), userHandler.List)
			users.DELETE("/:id", middleware.RequireAdmin(), userHandler.Delete)
			users.POST("/:id/block", middleware.RequireModerator(), userHandler.Block)
			users.POST("/:id/unblock", middleware.RequireModerator(), userHandler.Unblock)
		}

		// Project routes
		projects := api.Group("/projects")
		{
			projects.Use(middleware.OptionalAuthMiddleware())
			projects.GET("", projectHandler.List)
			projects.GET("/:id", projectHandler.Get)
			projects.POST("/:id/view", projectHandler.View)
			projects.GET("/:id/comments", commentHandler.List)

			// Protected project routes
			protectedProjects := projects.Group("")
			protectedProjects.Use(middleware.AuthMiddleware())
			{
				protectedProjects.POST("", projectHandler.Create)
				protectedProjects.PUT("/:id", projectHandler.Update)
				protectedProjects.DELETE("/:id", projectHandler.Delete)
				protectedProjects.POST("/:id/like", projectHandler.Like)
				protectedProjects.POST("/:id/comments", commentHandler.Create)

				// Moderator only
				protectedProjects.POST("/:id/block", middleware.RequireModerator(), projectHandler.Block)
				protectedProjects.POST("/:id/unblock", middleware.RequireModerator(), projectHandler.Unblock)
			}
		}

		// Comment routes (for deletion)
		comments := api.Group("/comments")
		comments.Use(middleware.AuthMiddleware())
		{
			comments.DELETE("/:id", commentHandler.Delete)
		}

		// Article routes
		articles := api.Group("/articles")
		{
			articles.Use(middleware.OptionalAuthMiddleware())
			articles.GET("", articleHandler.List)
			articles.GET("/:id", articleHandler.Get)
			articles.POST("/:id/view", articleHandler.View)

			// Protected article routes
			protectedArticles := articles.Group("")
			protectedArticles.Use(middleware.AuthMiddleware())
			{
				protectedArticles.POST("", articleHandler.Create)
				protectedArticles.PUT("/:id", articleHandler.Update)
				protectedArticles.DELETE("/:id", articleHandler.Delete)
			}
		}

		// Transaction routes
		transactions := api.Group("/transactions")
		{
			// Midtrans callback (public)
			transactions.POST("/callback", transactionHandler.Callback)

			// Protected transaction routes
			transactions.Use(middleware.AuthMiddleware())
			transactions.POST("", transactionHandler.Create)
			transactions.GET("", transactionHandler.List)
			transactions.GET("/check/:projectId", transactionHandler.CheckPurchase)

			// Admin only
			transactions.GET("/admin", middleware.RequireAdmin(), transactionHandler.AdminList)
		}

		// Category routes
		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.List)
			categories.GET("/:id", categoryHandler.Get)

			// Admin only
			categories.Use(middleware.AuthMiddleware(), middleware.RequireAdmin())
			categories.POST("", categoryHandler.Create)
			categories.PUT("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
		}

		// Upload routes
		upload := api.Group("/upload")
		upload.Use(middleware.AuthMiddleware())
		{
			upload.POST("", uploadHandler.Upload)
			upload.DELETE("/:filename", uploadHandler.Delete)
		}

		// Gamification routes
		gamification := api.Group("/gamification")
		{
			gamification.GET("/config", gamificationHandler.GetConfig)

			gamification.Use(middleware.AuthMiddleware())
			gamification.GET("/stats", gamificationHandler.GetStats)
		}
	}
}
