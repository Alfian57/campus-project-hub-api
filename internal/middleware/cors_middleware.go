package middleware

import (
	"time"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns configured CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig()

	return cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CORS.FrontendURL, "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
