package router

import (
	"github.com/campus-project-hub/api/internal/config"
	"github.com/campus-project-hub/api/internal/middleware"
	"github.com/gin-gonic/gin"

	// Swagger docs
	_ "github.com/campus-project-hub/api/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup initializes the Gin router with all routes and middleware
func Setup(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORSMiddleware())

	// Serve uploaded files
	router.Static("/uploads", cfg.Upload.Dir)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup API routes
	SetupV1Routes(router)

	return router
}
