package handlers

import (
	"strconv"
	"time"

	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/middleware"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/services"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ArticleHandler struct{}

func NewArticleHandler() *ArticleHandler {
	return &ArticleHandler{}
}

// List godoc
// @Summary      List articles
// @Description  Get paginated list of articles with optional filters
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        perPage query int false "Items per page" default(10)
// @Param        search query string false "Search by title or excerpt"
// @Param        category query string false "Filter by category"
// @Param        userId query string false "Filter by user ID"
// @Param        status query string false "Filter by status" default(published)
// @Success      200 {object} map[string]interface{} "Paginated articles list"
// @Router       /articles [get]
func (h *ArticleHandler) List(c *gin.Context) {
	db := database.GetDB()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	search := c.Query("search")
	category := c.Query("category")
	userID := c.Query("userId")
	status := c.DefaultQuery("status", "published")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	query := db.Model(&models.Article{}).Preload("User")

	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil || currentUser.Role == models.RoleUser {
		query = query.Where("status = ?", models.ArticleStatusPublished)
	} else if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Where("title ILIKE ? OR excerpt ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	query.Count(&total)

	var articles []models.Article
	query.Offset((page - 1) * perPage).Limit(perPage).Order("published_at DESC").Find(&articles)

	responses := make([]models.ArticleResponse, len(articles))
	for i, article := range articles {
		responses[i] = article.ToResponse()
	}

	utils.Paginated(c, responses, total, page, perPage)
}

// Get godoc
// @Summary      Get article by ID
// @Description  Get article details by article ID
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id path string true "Article ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Article details"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      404 {object} map[string]interface{} "Article not found"
// @Router       /articles/{id} [get]
func (h *ArticleHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()
	var article models.Article
	if err := db.Preload("User").First(&article, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Artikel tidak ditemukan")
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	if article.Status == models.ArticleStatusBlocked && (currentUser == nil || currentUser.Role == models.RoleUser) {
		utils.NotFound(c, "Artikel tidak ditemukan")
		return
	}

	utils.Success(c, article.ToResponse())
}

// CreateArticleInput for article creation
type CreateArticleInput struct {
	Title        string `json:"title" validate:"required,min=3,max=255"`
	Excerpt      string `json:"excerpt" validate:"required,max=500"`
	Content      string `json:"content" validate:"required"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Category     string `json:"category"`
	Status       string `json:"status" validate:"oneof=published draft"`
}

// Create godoc
// @Summary      Create article
// @Description  Create a new article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateArticleInput true "Article data"
// @Success      201 {object} map[string]interface{} "Created article"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /articles [post]
func (h *ArticleHandler) Create(c *gin.Context) {
	var input CreateArticleInput
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

	// Calculate reading time (approx 200 words per minute)
	wordCount := len(input.Content) / 5
	readingTime := wordCount / 200
	if readingTime < 1 {
		readingTime = 1
	}

	now := time.Now()
	article := models.Article{
		UserID:       currentUser.ID,
		Title:        input.Title,
		Excerpt:      &input.Excerpt,
		Content:      &input.Content,
		ThumbnailURL: &input.ThumbnailURL,
		Category:     &input.Category,
		ReadingTime:  readingTime,
		Status:       models.ArticleStatus(input.Status),
	}

	if input.Status == "published" {
		article.PublishedAt = &now
	}

	if err := db.Create(&article).Error; err != nil {
		utils.InternalServerError(c, "Gagal membuat artikel")
		return
	}

	// Add EXP for creating article
	services.AddUserExp(currentUser.ID, services.ExpCreateArticle)

	db.Preload("User").First(&article, "id = ?", article.ID)

	utils.Created(c, article.ToResponse())
}

// Update godoc
// @Summary      Update article
// @Description  Update an existing article (owner or admin only)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Article ID" format(uuid)
// @Param        request body CreateArticleInput true "Article data"
// @Success      200 {object} map[string]interface{} "Updated article"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Article not found"
// @Router       /articles/{id} [put]
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	var article models.Article
	if err := db.First(&article, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Artikel tidak ditemukan")
		return
	}

	if article.UserID != currentUser.ID && currentUser.Role != models.RoleAdmin {
		utils.Forbidden(c, "Tidak diizinkan mengubah artikel ini")
		return
	}

	var input CreateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	article.Title = input.Title
	article.Excerpt = &input.Excerpt
	article.Content = &input.Content
	article.ThumbnailURL = &input.ThumbnailURL
	article.Category = &input.Category
	article.Status = models.ArticleStatus(input.Status)

	// Update reading time
	wordCount := len(input.Content) / 5
	article.ReadingTime = wordCount / 200
	if article.ReadingTime < 1 {
		article.ReadingTime = 1
	}

	// Set published date if publishing for first time
	if input.Status == "published" && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	db.Save(&article)
	db.Preload("User").First(&article, "id = ?", article.ID)

	utils.Success(c, article.ToResponse())
}

// Delete godoc
// @Summary      Delete article
// @Description  Delete an article (owner or admin only)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Article ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Article deleted successfully"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Article not found"
// @Router       /articles/{id} [delete]
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	var article models.Article
	if err := db.First(&article, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Artikel tidak ditemukan")
		return
	}

	if article.UserID != currentUser.ID && currentUser.Role != models.RoleAdmin {
		utils.Forbidden(c, "Tidak diizinkan menghapus artikel ini")
		return
	}

	db.Delete(&article)
	utils.SuccessWithMessage(c, "Artikel berhasil dihapus", nil)
}

// View godoc
// @Summary      Record article view
// @Description  Increment article view count
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id path string true "Article ID" format(uuid)
// @Success      200 {object} map[string]interface{} "View recorded"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Router       /articles/{id}/view [post]
func (h *ArticleHandler) View(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()
	db.Model(&models.Article{}).Where("id = ?", id).Update("views", db.Raw("views + 1"))

	var article models.Article
	if err := db.First(&article, "id = ?", id).Error; err == nil {
		services.AddUserExp(article.UserID, services.ExpArticleViewed)
	}

	utils.SuccessWithMessage(c, "View recorded", nil)
}
