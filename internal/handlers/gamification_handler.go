package handlers

import (
	"github.com/campus-project-hub/api/internal/middleware"
	"github.com/campus-project-hub/api/internal/services"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/gin-gonic/gin"
)

type GamificationHandler struct{}

func NewGamificationHandler() *GamificationHandler {
	return &GamificationHandler{}
}

// GetConfig godoc
// @Summary      Get gamification config
// @Description  Get gamification configuration (EXP values, levels)
// @Tags         gamification
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "Gamification config"
// @Router       /gamification/config [get]
func (h *GamificationHandler) GetConfig(c *gin.Context) {
	config := services.GetGamificationConfig()
	utils.Success(c, config)
}

// GetStats godoc
// @Summary      Get user stats
// @Description  Get current user's gamification stats
// @Tags         gamification
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "User stats"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /gamification/stats [get]
func (h *GamificationHandler) GetStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.Unauthorized(c, "Tidak terautentikasi")
		return
	}

	stats := services.GetUserGamificationStats(user.TotalExp)
	utils.Success(c, stats)
}
