package middleware

import (
	"strings"

	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
	UserIDContextKey    = "userId"
)

// AuthMiddleware validates JWT token and sets user in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			utils.Unauthorized(c, "Token tidak ditemukan")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			utils.Unauthorized(c, "Format token tidak valid")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.Unauthorized(c, "Token tidak valid atau sudah kedaluwarsa")
			c.Abort()
			return
		}

		// Fetch user from database
		var user models.User
		if err := database.GetDB().First(&user, "id = ?", claims.UserID).Error; err != nil {
			utils.Unauthorized(c, "User tidak ditemukan")
			c.Abort()
			return
		}

		// Check if user is blocked
		if user.Status == models.StatusBlocked {
			utils.Forbidden(c, "Akun Anda telah diblokir")
			c.Abort()
			return
		}

		c.Set(UserContextKey, &user)
		c.Set(UserIDContextKey, user.ID)
		c.Next()
	}
}

// OptionalAuthMiddleware tries to authenticate but doesn't require it
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		var user models.User
		if err := database.GetDB().First(&user, "id = ?", claims.UserID).Error; err == nil {
			if user.Status != models.StatusBlocked {
				c.Set(UserContextKey, &user)
				c.Set(UserIDContextKey, user.ID)
			}
		}

		c.Next()
	}
}

// GetCurrentUser retrieves the authenticated user from context
func GetCurrentUser(c *gin.Context) *models.User {
	if user, exists := c.Get(UserContextKey); exists {
		return user.(*models.User)
	}
	return nil
}

// GetCurrentUserID retrieves the authenticated user ID from context
func GetCurrentUserID(c *gin.Context) uuid.UUID {
	if userID, exists := c.Get(UserIDContextKey); exists {
		return userID.(uuid.UUID)
	}
	return uuid.Nil
}
