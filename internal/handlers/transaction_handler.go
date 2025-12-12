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

type TransactionHandler struct{}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

// CreateTransactionInput for creating a transaction
type CreateTransactionInput struct {
	ProjectID string `json:"projectId" validate:"required,uuid"`
}

// Create godoc
// @Summary      Create transaction
// @Description  Initiate a payment transaction for a project
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateTransactionInput true "Transaction data"
// @Success      201 {object} map[string]interface{} "Snap token and redirect URL"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	var input CreateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	if err := utils.Validate(&input); err != nil {
		errors := utils.FormatValidationErrors(err)
		c.JSON(400, gin.H{"success": false, "errors": errors})
		return
	}

	projectID, _ := uuid.Parse(input.ProjectID)
	currentUser := middleware.GetCurrentUser(c)

	snapResp, transaction, err := services.CreateSnapTransaction(
		projectID,
		currentUser.ID,
		currentUser.Email,
		currentUser.Name,
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, gin.H{
		"token":         snapResp.Token,
		"redirectUrl":   snapResp.RedirectURL,
		"transactionId": transaction.ID,
	})
}

// Callback godoc
// @Summary      Midtrans callback
// @Description  Handle Midtrans webhook notification
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        request body services.MidtransNotification true "Midtrans notification"
// @Success      200 {object} map[string]interface{} "OK"
// @Failure      400 {object} map[string]interface{} "Invalid notification"
// @Router       /transactions/callback [post]
func (h *TransactionHandler) Callback(c *gin.Context) {
	var notification services.MidtransNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		utils.BadRequest(c, "Data tidak valid")
		return
	}

	if err := services.HandleMidtransNotification(&notification); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "OK", nil)
}

// List godoc
// @Summary      List user transactions
// @Description  Get paginated list of user's transactions
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        perPage query int false "Items per page" default(10)
// @Param        type query string false "Filter type (all, purchases, sales)" default(all)
// @Success      200 {object} map[string]interface{} "Paginated transactions"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	transactionType := c.DefaultQuery("type", "all") // all, purchases, sales

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	query := db.Model(&models.Transaction{}).Preload("Project").Preload("Buyer")

	switch transactionType {
	case "purchases":
		query = query.Where("buyer_id = ?", currentUser.ID)
	case "sales":
		query = query.Where("seller_id = ?", currentUser.ID)
	default:
		query = query.Where("buyer_id = ? OR seller_id = ?", currentUser.ID, currentUser.ID)
	}

	var total int64
	query.Count(&total)

	var transactions []models.Transaction
	query.Offset((page - 1) * perPage).Limit(perPage).Order("created_at DESC").Find(&transactions)

	responses := make([]models.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = tx.ToResponse()
	}

	utils.Paginated(c, responses, total, page, perPage)
}

// AdminList godoc
// @Summary      Admin list transactions
// @Description  Get all transactions (admin only)
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        perPage query int false "Items per page" default(10)
// @Param        status query string false "Filter by status"
// @Success      200 {object} map[string]interface{} "Paginated transactions"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Router       /transactions/admin [get]
func (h *TransactionHandler) AdminList(c *gin.Context) {
	db := database.GetDB()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	query := db.Model(&models.Transaction{}).Preload("Project").Preload("Buyer").Preload("Seller")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var transactions []models.Transaction
	query.Offset((page - 1) * perPage).Limit(perPage).Order("created_at DESC").Find(&transactions)

	responses := make([]models.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = tx.ToResponse()
	}

	utils.Paginated(c, responses, total, page, perPage)
}

// CheckPurchase godoc
// @Summary      Check purchase status
// @Description  Check if user has purchased a project
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectId path string true "Project ID" format(uuid)
// @Success      200 {object} map[string]interface{} "Purchase status"
// @Failure      400 {object} map[string]interface{} "Invalid ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /transactions/check/{projectId} [get]
func (h *TransactionHandler) CheckPurchase(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		utils.BadRequest(c, "Project ID tidak valid")
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	db := database.GetDB()

	var transaction models.Transaction
	err = db.Where("project_id = ? AND buyer_id = ? AND status = ?",
		projectID, currentUser.ID, models.TransactionStatusSuccess).First(&transaction).Error

	utils.Success(c, gin.H{
		"purchased": err == nil,
	})
}
