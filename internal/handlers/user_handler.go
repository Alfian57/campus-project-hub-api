package handlers

import (
	"strconv"

	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/middleware"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/services"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// List godoc
// @Summary      List users
// @Description  Get paginated list of users (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        perPage query int false "Items per page" default(10)
// @Param        search query string false "Search by name or email"
// @Param        role query string false "Filter by role (user, admin, moderator)"
// @Param        status query string false "Filter by status (active, blocked)"
// @Success      200 {object} map[string]interface{} "Paginated users list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Router       /users [get]
func (h *UserHandler) List(c *gin.Context) {
	db := database.GetDB()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	query := db.Model(&models.User{})

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var users []models.User
	query.Offset((page - 1) * perPage).Limit(perPage).Order("created_at DESC").Find(&users)

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	utils.Paginated(c, responses, total, page, perPage)
}

// Get godoc
// @Summary      Get user by ID
// @Description  Get user profile by user ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Success      200 {object} map[string]interface{} "User profile with gamification stats"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Router       /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	user, err := services.GetUserByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	stats := services.GetUserGamificationStats(user.TotalExp)

	utils.Success(c, gin.H{
		"user":         user.ToResponse(),
		"gamification": stats,
	})
}

// Update godoc
// @Summary      Update user profile
// @Description  Update user profile (self or admin)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID" format(uuid)
// @Param        request body services.UpdateUserInput true "User profile data"
// @Success      200 {object} map[string]interface{} "Updated user profile"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	currentUser := middleware.GetCurrentUser(c)

	// Only allow self-update or admin update
	if currentUser.ID != id && currentUser.Role != models.RoleAdmin {
		utils.Forbidden(c, "Tidak diizinkan mengubah profil user lain")
		return
	}

	var input services.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	user, err := services.UpdateUser(id, &input)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, user.ToResponse())
}

// Delete godoc
// @Summary      Delete user
// @Description  Delete a user (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID" format(uuid)
// @Success      200 {object} map[string]interface{} "User deleted successfully"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()
	if err := db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		utils.InternalServerError(c, "Gagal menghapus user")
		return
	}

	utils.SuccessWithMessage(c, "User berhasil dihapus", nil)
}

// Block godoc
// @Summary      Block user
// @Description  Block a user (moderator/admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID" format(uuid)
// @Param        request body object{reason=string} false "Block reason"
// @Success      200 {object} map[string]interface{} "User blocked successfully"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Router       /users/{id}/block [post]
func (h *UserHandler) Block(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&input)

	db := database.GetDB()
	currentUser := middleware.GetCurrentUser(c)

	var user models.User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "User tidak ditemukan")
		return
	}

	// Update user status
	user.Status = models.StatusBlocked
	db.Save(&user)

	// Create block record
	blockRecord := models.BlockRecord{
		TargetType: models.TargetTypeUser,
		TargetID:   user.ID,
		TargetName: user.Name,
		Reason:     &input.Reason,
		BlockedBy:  currentUser.ID,
	}
	db.Create(&blockRecord)

	utils.SuccessWithMessage(c, "User berhasil diblokir", nil)
}

// Unblock godoc
// @Summary      Unblock user
// @Description  Unblock a user (moderator/admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID" format(uuid)
// @Success      200 {object} map[string]interface{} "User unblocked successfully"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Router       /users/{id}/unblock [post]
func (h *UserHandler) Unblock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()

	var user models.User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "User tidak ditemukan")
		return
	}

	user.Status = models.StatusActive
	db.Save(&user)

	// Remove block record
	db.Delete(&models.BlockRecord{}, "target_type = ? AND target_id = ?", models.TargetTypeUser, id)

	utils.SuccessWithMessage(c, "User berhasil dibuka blokirnya", nil)
}

// Leaderboard godoc
// @Summary      Get leaderboard
// @Description  Get top users by EXP
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        limit query int false "Number of users to return" default(10)
// @Success      200 {object} map[string]interface{} "Leaderboard entries"
// @Router       /users/leaderboard [get]
func (h *UserHandler) Leaderboard(c *gin.Context) {
	db := database.GetDB()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var users []models.User
	db.Where("status = ?", models.StatusActive).
		Order("total_exp DESC").
		Limit(limit).
		Find(&users)

	type LeaderboardEntry struct {
		Rank       int                 `json:"rank"`
		User       models.UserResponse `json:"user"`
		TotalExp   int                 `json:"totalExp"`
		Level      int                 `json:"level"`
		LevelTitle string              `json:"levelTitle"`
	}

	entries := make([]LeaderboardEntry, len(users))
	for i, user := range users {
		stats := services.GetUserGamificationStats(user.TotalExp)
		entries[i] = LeaderboardEntry{
			Rank:       i + 1,
			User:       user.ToResponse(),
			TotalExp:   user.TotalExp,
			Level:      stats.Level,
			LevelTitle: stats.LevelTitle,
		}
	}

	utils.Success(c, gin.H{
		"leaderboard": entries,
	})
}
