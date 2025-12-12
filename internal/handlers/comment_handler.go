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

type CommentHandler struct{}

func NewCommentHandler() *CommentHandler {
	return &CommentHandler{}
}

// List godoc
// @Summary      List project comments
// @Description  Get paginated list of comments for a project
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID" format(uuid)
// @Param        page query int false "Page number" default(1)
// @Param        perPage query int false "Items per page" default(20)
// @Success      200 {object} map[string]interface{} "Paginated comments list"
// @Failure      400 {object} map[string]interface{} "Invalid project ID"
// @Router       /projects/{id}/comments [get]
func (h *CommentHandler) List(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Project ID tidak valid")
		return
	}

	db := database.GetDB()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int64
	db.Model(&models.Comment{}).Where("project_id = ?", projectID).Count(&total)

	var comments []models.Comment
	db.Preload("User").
		Where("project_id = ?", projectID).
		Offset((page - 1) * perPage).
		Limit(perPage).
		Order("created_at DESC").
		Find(&comments)

	responses := make([]models.CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = comment.ToResponse()
	}

	utils.Paginated(c, responses, total, page, perPage)
}

// CreateCommentInput for creating a comment
type CreateCommentInput struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// Create godoc
// @Summary      Create comment
// @Description  Add a comment to a project
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project ID" format(uuid)
// @Param        request body CreateCommentInput true "Comment data"
// @Success      201 {object} map[string]interface{} "Created comment"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /projects/{id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Project ID tidak valid")
		return
	}

	var input CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	if err := utils.Validate(&input); err != nil {
		errors := utils.FormatValidationErrors(err)
		c.JSON(400, gin.H{"success": false, "errors": errors})
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	// Check if project exists
	var project models.Project
	if err := db.First(&project, "id = ?", projectID).Error; err != nil {
		utils.NotFound(c, "Project tidak ditemukan")
		return
	}

	comment := models.Comment{
		UserID:    currentUser.ID,
		ProjectID: projectID,
		Content:   input.Content,
	}

	if err := db.Create(&comment).Error; err != nil {
		utils.InternalServerError(c, "Gagal membuat komentar")
		return
	}

	// Add EXP to project owner
	if project.UserID != currentUser.ID {
		services.AddUserExp(project.UserID, services.ExpReceiveComment)
	}

	db.Preload("User").First(&comment, "id = ?", comment.ID)

	utils.Created(c, comment.ToResponse())
}

// Delete godoc
// @Summary      Delete comment
// @Description  Delete a comment (owner, project owner, admin, or moderator)
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Comment ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Comment deleted successfully"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Comment not found"
// @Router       /comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Comment ID tidak valid")
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	var comment models.Comment
	if err := db.Preload("Project").First(&comment, "id = ?", commentID).Error; err != nil {
		utils.NotFound(c, "Komentar tidak ditemukan")
		return
	}

	// Allow delete by: comment owner, project owner, admin, or moderator
	canDelete := comment.UserID == currentUser.ID ||
		comment.Project.UserID == currentUser.ID ||
		currentUser.Role == models.RoleAdmin ||
		currentUser.Role == models.RoleModerator

	if !canDelete {
		utils.Forbidden(c, "Tidak diizinkan menghapus komentar ini")
		return
	}

	db.Delete(&comment)
	utils.SuccessWithMessage(c, "Komentar berhasil dihapus", nil)
}
