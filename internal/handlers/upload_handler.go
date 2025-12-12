package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// Upload godoc
// @Summary      Upload file
// @Description  Upload a file (image, PDF, or ZIP)
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "File to upload"
// @Success      200 {object} map[string]interface{} "Upload result"
// @Failure      400 {object} map[string]interface{} "Invalid file"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Upload failed"
// @Router       /upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	cfg := config.GetConfig()

	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "File tidak ditemukan")
		return
	}

	// Check file size
	if file.Size > cfg.Upload.MaxSize {
		utils.BadRequest(c, fmt.Sprintf("Ukuran file maksimal %dMB", cfg.Upload.MaxSize/1024/1024))
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".pdf":  true,
		".zip":  true,
	}
	if !allowedExts[ext] {
		utils.BadRequest(c, "Tipe file tidak diizinkan")
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	uploadPath := filepath.Join(cfg.Upload.Dir, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.Upload.Dir, 0755); err != nil {
		utils.InternalServerError(c, "Gagal menyiapkan direktori upload")
		return
	}

	// Save file
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		utils.InternalServerError(c, "Gagal menyimpan file")
		return
	}

	// Return the URL
	fileURL := fmt.Sprintf("/uploads/%s", filename)

	utils.Success(c, gin.H{
		"filename": filename,
		"url":      fileURL,
		"size":     file.Size,
	})
}

// Delete godoc
// @Summary      Delete file
// @Description  Delete an uploaded file
// @Tags         upload
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        filename path string true "Filename"
// @Success      200 {object} map[string]interface{} "File deleted"
// @Failure      400 {object} map[string]interface{} "Invalid filename"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "File not found"
// @Failure      500 {object} map[string]interface{} "Delete failed"
// @Router       /upload/{filename} [delete]
func (h *UploadHandler) Delete(c *gin.Context) {
	cfg := config.GetConfig()
	filename := c.Param("filename")

	// Sanitize filename
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		utils.BadRequest(c, "Nama file tidak valid")
		return
	}

	filePath := filepath.Join(cfg.Upload.Dir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.NotFound(c, "File tidak ditemukan")
		return
	}

	if err := os.Remove(filePath); err != nil {
		utils.InternalServerError(c, "Gagal menghapus file")
		return
	}

	utils.SuccessWithMessage(c, "File berhasil dihapus", nil)
}
