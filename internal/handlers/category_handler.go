package handlers

import (
	"regexp"
	"strings"

	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

// List godoc
// @Summary      List categories
// @Description  Get all categories with project counts
// @Tags         categories
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "Categories list"
// @Router       /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	db := database.GetDB()

	var categories []models.Category
	db.Order("name ASC").Find(&categories)

	// Count projects for each category
	for i := range categories {
		var count int64
		db.Model(&models.Project{}).
			Where("category_id = ? AND status = ?", categories[i].ID, models.ProjectStatusPublished).
			Count(&count)
		categories[i].ProjectCount = int(count)
	}

	utils.Success(c, categories)
}

// Get godoc
// @Summary      Get category by ID
// @Description  Get category details by category ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path string true "Category ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Category details"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      404 {object} map[string]interface{} "Category not found"
// @Router       /categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()
	var category models.Category
	if err := db.First(&category, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Kategori tidak ditemukan")
		return
	}

	var count int64
	db.Model(&models.Project{}).
		Where("category_id = ? AND status = ?", category.ID, models.ProjectStatusPublished).
		Count(&count)
	category.ProjectCount = int(count)

	utils.Success(c, category)
}

// CreateCategoryInput for category creation
type CreateCategoryInput struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}

// Create godoc
// @Summary      Create category
// @Description  Create a new category (admin only)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateCategoryInput true "Category data"
// @Success      201 {object} map[string]interface{} "Created category"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Router       /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var input CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	if err := utils.Validate(&input); err != nil {
		errors := utils.FormatValidationErrors(err)
		c.JSON(400, gin.H{"success": false, "errors": errors})
		return
	}

	db := database.GetDB()

	slug := generateSlug(input.Name)

	// Check if slug exists
	var existing models.Category
	if err := db.Where("slug = ?", slug).First(&existing).Error; err == nil {
		utils.BadRequest(c, "Kategori dengan nama serupa sudah ada")
		return
	}

	category := models.Category{
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		Color:       input.Color,
	}

	if err := db.Create(&category).Error; err != nil {
		utils.InternalServerError(c, "Gagal membuat kategori")
		return
	}

	utils.Created(c, category)
}

// Update godoc
// @Summary      Update category
// @Description  Update a category (admin only)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Category ID" format(uuid)
// @Param        request body CreateCategoryInput true "Category data"
// @Success      200 {object} map[string]interface{} "Updated category"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Category not found"
// @Router       /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()

	var category models.Category
	if err := db.First(&category, "id = ?", id).Error; err != nil {
		utils.NotFound(c, "Kategori tidak ditemukan")
		return
	}

	var input CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	newSlug := generateSlug(input.Name)
	if newSlug != category.Slug {
		var existing models.Category
		if err := db.Where("slug = ? AND id != ?", newSlug, id).First(&existing).Error; err == nil {
			utils.BadRequest(c, "Kategori dengan nama serupa sudah ada")
			return
		}
		category.Slug = newSlug
	}

	category.Name = input.Name
	category.Description = input.Description
	category.Color = input.Color

	db.Save(&category)

	utils.Success(c, category)
}

// Delete godoc
// @Summary      Delete category
// @Description  Delete a category (admin only)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Category ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Category deleted"
// @Failure      400 {object} map[string]interface{} "Invalid ID or has projects"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID tidak valid")
		return
	}

	db := database.GetDB()

	// Check if category has projects
	var count int64
	db.Model(&models.Project{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		utils.BadRequest(c, "Tidak dapat menghapus kategori yang memiliki project")
		return
	}

	if err := db.Delete(&models.Category{}, "id = ?", id).Error; err != nil {
		utils.InternalServerError(c, "Gagal menghapus kategori")
		return
	}

	utils.SuccessWithMessage(c, "Kategori berhasil dihapus", nil)
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	return slug
}
