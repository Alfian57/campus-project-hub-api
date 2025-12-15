package middleware

import (
	"strings"
	"time"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns configured CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig()

	origins := []string{"http://localhost:3000"}
	if cfg.CORS.FrontendURL != "" {
		configuredOrigins := strings.Split(cfg.CORS.FrontendURL, ",")
		for _, origin := range configuredOrigins {
			trimmedOrigin := strings.TrimSpace(origin)
			if trimmedOrigin != "" {
				origins = append(origins, trimmedOrigin)
			}
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
