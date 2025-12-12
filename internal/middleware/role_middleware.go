package middleware

import (
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/gin-gonic/gin"
)

// RequireRole middleware checks if user has specific role
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			utils.Unauthorized(c, "Autentikasi diperlukan")
			c.Abort()
			return
		}

		// Admin has access to everything
		if user.Role == models.RoleAdmin {
			c.Next()
			return
		}

		// Check if user has required role
		for _, role := range roles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "Akses ditolak")
		c.Abort()
	}
}

// RequireAdmin middleware checks if user is admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireModerator middleware checks if user is admin or moderator
func RequireModerator() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleModerator)
}
