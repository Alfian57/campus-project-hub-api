package services

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/google/uuid"
)

// MidtransSnapRequest for creating transaction
type MidtransSnapRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	CreditCard         CreditCard         `json:"credit_card"`
	ItemDetails        []ItemDetail       `json:"item_details"`
	CustomerDetails    CustomerDetails    `json:"customer_details"`
}

type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int    `json:"gross_amount"`
}

type CreditCard struct {
	Secure bool `json:"secure"`
}

type ItemDetail struct {
	ID       string `json:"id"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
}

// MidtransSnapResponse from Snap API
type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

// MidtransNotification from webhook
type MidtransNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
}

// CreateSnapTransaction creates Midtrans Snap transaction
func CreateSnapTransaction(projectID, buyerID uuid.UUID, buyerEmail, buyerName string) (*MidtransSnapResponse, *models.Transaction, error) {
	cfg := config.GetConfig()
	db := database.GetDB()

	// Get project
	var project models.Project
	if err := db.Preload("User").First(&project, "id = ?", projectID).Error; err != nil {
		return nil, nil, fmt.Errorf("project tidak ditemukan")
	}

	if project.Type != models.ProjectTypePaid {
		return nil, nil, fmt.Errorf("project ini gratis")
	}

	// Check if already purchased
	var existingTx models.Transaction
	if err := db.Where("project_id = ? AND buyer_id = ? AND status = ?",
		projectID, buyerID, models.TransactionStatusSuccess).First(&existingTx).Error; err == nil {
		return nil, nil, fmt.Errorf("Anda sudah membeli project ini")
	}

	// Create transaction record
	orderID := fmt.Sprintf("PURCHASE-%s-%d", projectID.String()[:8], time.Now().Unix())
	transaction := models.Transaction{
		ProjectID:       projectID,
		BuyerID:         buyerID,
		SellerID:        project.UserID,
		Amount:          project.Price,
		Status:          models.TransactionStatusPending,
		MidtransOrderID: &orderID,
	}

	if err := db.Create(&transaction).Error; err != nil {
		return nil, nil, fmt.Errorf("gagal membuat transaksi: %w", err)
	}

	// Create Snap request
	snapReq := MidtransSnapRequest{
		TransactionDetails: TransactionDetails{
			OrderID:     orderID,
			GrossAmount: project.Price,
		},
		CreditCard: CreditCard{
			Secure: true,
		},
		ItemDetails: []ItemDetail{
			{
				ID:       projectID.String(),
				Price:    project.Price,
				Quantity: 1,
				Name:     truncateString(project.Title, 50),
			},
		},
		CustomerDetails: CustomerDetails{
			FirstName: buyerName,
			Email:     buyerEmail,
		},
	}

	// Call Midtrans Snap API
	baseURL := "https://app.sandbox.midtrans.com"
	if cfg.Midtrans.IsProduction {
		baseURL = "https://app.midtrans.com"
	}

	jsonData, _ := json.Marshal(snapReq)
	req, _ := http.NewRequest("POST", baseURL+"/snap/v1/transactions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(cfg.Midtrans.ServerKey, "")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal menghubungi Midtrans: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var snapResp MidtransSnapResponse
	if err := json.Unmarshal(body, &snapResp); err != nil {
		return nil, nil, fmt.Errorf("gagal memproses respons Midtrans: %w", err)
	}

	if snapResp.Token == "" {
		return nil, nil, fmt.Errorf("gagal membuat transaksi Midtrans")
	}

	return &snapResp, &transaction, nil
}

// HandleMidtransNotification processes webhook notification
func HandleMidtransNotification(notification *MidtransNotification) error {
	cfg := config.GetConfig()
	db := database.GetDB()

	// Verify signature
	signatureData := notification.OrderID + notification.StatusCode + notification.GrossAmount + cfg.Midtrans.ServerKey
	hash := sha512.Sum512([]byte(signatureData))
	expectedSignature := hex.EncodeToString(hash[:])

	if notification.SignatureKey != expectedSignature {
		return fmt.Errorf("signature tidak valid")
	}

	// Find transaction
	var transaction models.Transaction
	if err := db.Where("midtrans_order_id = ?", notification.OrderID).First(&transaction).Error; err != nil {
		return fmt.Errorf("transaksi tidak ditemukan")
	}

	// Update transaction status
	transaction.MidtransTransactionID = &notification.TransactionID

	switch notification.TransactionStatus {
	case "capture", "settlement":
		if notification.FraudStatus == "accept" || notification.FraudStatus == "" {
			transaction.Status = models.TransactionStatusSuccess

			// Add EXP to buyer and seller
			AddUserExp(transaction.BuyerID, ExpBuyProject)
			AddUserExp(transaction.SellerID, ExpSellProject)
		}
	case "pending":
		transaction.Status = models.TransactionStatusPending
	case "deny", "cancel", "expire":
		transaction.Status = models.TransactionStatusFailed
	}

	return db.Save(&transaction).Error
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
