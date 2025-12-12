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
	"github.com/lib/pq"
)

type ProjectHandler struct{}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{}
}

// List godoc
// @Summary      List projects
// @Description  Get paginated list of projects with optional filters
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        perPage query int false "Items per page" default(12)
// @Param        search query string false "Search by title or description"
// @Param        type query string false "Filter by type (free, paid)"
// @Param        categoryId query string false "Filter by category ID"
// @Param        userId query string false "Filter by user ID"
// @Param        status query string false "Filter by status (published, draft, blocked)" default(published)
// @Success      200 {object} map[string]interface{} "Paginated projects list"
// @Router       /projects [get]
func (h *ProjectHandler) List(c *gin.Context) {
	db := database.GetDB()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "12"))
	search := c.Query("search")
	projectType := c.Query("type")
	categoryID := c.Query("categoryId")
	userID := c.Query("userId")
	status := c.DefaultQuery("status", "published")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 12
	}

	query := db.Model(&models.Project{}).Preload("User").Preload("Images")

	// Only show published projects to non-admins
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil || currentUser.Role == models.RoleUser {
		query = query.Where("status = ?", models.ProjectStatusPublished)
	} else if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if projectType != "" {
		query = query.Where("type = ?", projectType)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	query.Count(&total)

	var projects []models.Project
	query.Offset((page - 1) * perPage).Limit(perPage).Order("created_at DESC").Find(&projects)

	responses := make([]models.ProjectResponse, len(projects))
	for i, project := range projects {
		var commentCount int64
		db.Model(&models.Comment{}).Where("project_id = ?", project.ID).Count(&commentCount)
		responses[i] = project.ToResponse(int(commentCount))
	}

	utils.Paginated(c, responses, total, page, perPage)
}

// Get godoc
// @Summary      Get project by ID
// @Description  Get project details by project ID
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Project details"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Router       /projects/{id} [get]
func (h *ProjectHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()
	var project models.Project
	if err := db.Preload("User").Preload("Images").Preload("Category").First(&project, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Project tidak ditemukan")
		return
	}

	// Check if blocked
	currentUser := middleware.GetCurrentUser(c)
	if project.Status == models.ProjectStatusBlocked && (currentUser == nil || currentUser.Role == models.RoleUser) {
		utils.NotFound(c, "Project tidak ditemukan")
		return
	}

	var commentCount int64
	db.Model(&models.Comment{}).Where("project_id = ?", project.ID).Count(&commentCount)

	utils.Success(c, project.ToResponse(int(commentCount)))
}

// CreateProjectInput for project creation
type CreateProjectInput struct {
	Title        string   `json:"title" validate:"required,min=3,max=255"`
	Description  string   `json:"description" validate:"required"`
	ThumbnailURL string   `json:"thumbnailUrl"`
	Images       []string `json:"images"`
	TechStack    []string `json:"techStack"`
	GithubURL    string   `json:"githubUrl" validate:"omitempty,url"`
	DemoURL      string   `json:"demoUrl" validate:"omitempty,url"`
	Type         string   `json:"type" validate:"required,oneof=free paid"`
	Price        int      `json:"price" validate:"omitempty,min=0"`
	CategoryID   string   `json:"categoryId"`
}

// Create godoc
// @Summary      Create project
// @Description  Create a new project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateProjectInput true "Project data"
// @Success      201 {object} map[string]interface{} "Created project"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /projects [post]
func (h *ProjectHandler) Create(c *gin.Context) {
	var input CreateProjectInput
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

	project := models.Project{
		UserID:       currentUser.ID,
		Title:        input.Title,
		Description:  &input.Description,
		ThumbnailURL: &input.ThumbnailURL,
		TechStack:    pq.StringArray(input.TechStack),
		GithubURL:    &input.GithubURL,
		DemoURL:      &input.DemoURL,
		Type:         models.ProjectType(input.Type),
		Price:        input.Price,
		Status:       models.ProjectStatusPublished,
	}

	if input.CategoryID != "" {
		catID, _ := uuid.Parse(input.CategoryID)
		project.CategoryID = &catID
	}

	if err := db.Create(&project).Error; err != nil {
		utils.InternalServerError(c, "Gagal membuat project")
		return
	}

	// Create project images
	for i, imgURL := range input.Images {
		img := models.ProjectImage{
			ProjectID: project.ID,
			ImageURL:  imgURL,
			SortOrder: i,
		}
		db.Create(&img)
	}

	// Add EXP for creating project
	services.AddUserExp(currentUser.ID, services.ExpCreateProject)

	// Reload with relations
	db.Preload("User").Preload("Images").First(&project, "id = ?", project.ID)

	utils.Created(c, project.ToResponse(0))
}

// Update godoc
// @Summary      Update project
// @Description  Update an existing project (owner or admin only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project ID" format(uuid)
// @Param        request body CreateProjectInput true "Project data"
// @Success      200 {object} map[string]interface{} "Updated project"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Router       /projects/{id} [put]
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	var project models.Project
	if err := db.First(&project, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Project tidak ditemukan")
		return
	}

	// Only owner or admin can update
	if project.UserID != currentUser.ID && currentUser.Role != models.RoleAdmin {
		utils.Forbidden(c, "Tidak diizinkan mengubah project ini")
		return
	}

	var input CreateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	project.Title = input.Title
	project.Description = &input.Description
	project.ThumbnailURL = &input.ThumbnailURL
	project.TechStack = pq.StringArray(input.TechStack)
	project.GithubURL = &input.GithubURL
	project.DemoURL = &input.DemoURL
	project.Type = models.ProjectType(input.Type)
	project.Price = input.Price

	if input.CategoryID != "" {
		catID, _ := uuid.Parse(input.CategoryID)
		project.CategoryID = &catID
	}

	db.Save(&project)

	// Update images
	db.Delete(&models.ProjectImage{}, "project_id = ?", id)
	for i, imgURL := range input.Images {
		img := models.ProjectImage{
			ProjectID: project.ID,
			ImageURL:  imgURL,
			SortOrder: i,
		}
		db.Create(&img)
	}

	db.Preload("User").Preload("Images").First(&project, "id = ?", project.ID)

	var commentCount int64
	db.Model(&models.Comment{}).Where("project_id = ?", project.ID).Count(&commentCount)

	utils.Success(c, project.ToResponse(int(commentCount)))
}

// Delete godoc
// @Summary      Delete project
// @Description  Delete a project (owner or admin only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Project deleted successfully"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Router       /projects/{id} [delete]
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	var project models.Project
	if err := db.First(&project, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Project tidak ditemukan")
		return
	}

	// Only owner or admin can delete
	if project.UserID != currentUser.ID && currentUser.Role != models.RoleAdmin {
		utils.Forbidden(c, "Tidak diizinkan menghapus project ini")
		return
	}

	db.Delete(&project)
	utils.SuccessWithMessage(c, "Project berhasil dihapus", nil)
}

// Like godoc
// @Summary      Toggle project like
// @Description  Like or unlike a project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Like status and count"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Router       /projects/{id}/like [post]
func (h *ProjectHandler) Like(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	var project models.Project
	if err := db.First(&project, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Project tidak ditemukan")
		return
	}

	// Check if already liked
	var existingLike models.ProjectLike
	if err := db.Where("user_id = ? AND project_id = ?", currentUser.ID, id).First(&existingLike).Error; err == nil {
		// Unlike
		db.Delete(&existingLike)
		db.Model(&project).Update("likes", project.Likes-1)
		utils.Success(c, gin.H{"liked": false, "likes": project.Likes - 1})
		return
	}

	// Like
	like := models.ProjectLike{
		UserID:    currentUser.ID,
		ProjectID: id,
	}
	db.Create(&like)
	db.Model(&project).Update("likes", project.Likes+1)

	// Add EXP to project owner
	if project.UserID != currentUser.ID {
		services.AddUserExp(project.UserID, services.ExpReceiveLike)
	}

	utils.Success(c, gin.H{"liked": true, "likes": project.Likes + 1})
}

// View godoc
// @Summary      Record project view
// @Description  Increment project view count
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID" format(uuid)
// @Success      200 {object} map[string]interface{} "View recorded"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Router       /projects/{id}/view [post]
func (h *ProjectHandler) View(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()
	db.Model(&models.Project{}).Where("id = ?", id).Update("views", db.Raw("views + 1"))

	// Add EXP to owner (could add daily cap logic here)
	var project models.Project
	if err := db.First(&project, "id = ?", id).Error; err == nil {
		services.AddUserExp(project.UserID, services.ExpProjectViewed)
	}

	utils.SuccessWithMessage(c, "View recorded", nil)
}

// Block godoc
// @Summary      Block project
// @Description  Block a project (moderator/admin only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project ID" format(uuid)
// @Param        request body object{reason=string} false "Block reason"
// @Success      200 {object} map[string]interface{} "Project blocked successfully"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Router       /projects/{id}/block [post]
func (h *ProjectHandler) Block(c *gin.Context) {
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

	var project models.Project
	if err := db.Preload("User").First(&project, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Project tidak ditemukan")
		return
	}

	project.Status = models.ProjectStatusBlocked
	db.Save(&project)

	blockRecord := models.BlockRecord{
		TargetType: models.TargetTypeProject,
		TargetID:   project.ID,
		TargetName: project.Title,
		Reason:     &input.Reason,
		BlockedBy:  currentUser.ID,
	}
	db.Create(&blockRecord)

	utils.SuccessWithMessage(c, "Project berhasil diblokir", nil)
}

// Unblock godoc
// @Summary      Unblock project
// @Description  Unblock a project (moderator/admin only)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Project unblocked successfully"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Project not found"
// @Router       /projects/{id}/unblock [post]
func (h *ProjectHandler) Unblock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()

	var project models.Project
	if err := db.First(&project, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Project tidak ditemukan")
		return
	}

	project.Status = models.ProjectStatusPublished
	db.Save(&project)

	db.Delete(&models.BlockRecord{}, "target_type = ? AND target_id = ?", models.TargetTypeProject, id)

	utils.SuccessWithMessage(c, "Project berhasil dibuka blokirnya", nil)
}
