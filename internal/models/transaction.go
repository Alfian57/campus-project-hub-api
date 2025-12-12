package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)

type Transaction struct {
	ID                    uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID             uuid.UUID         `gorm:"type:uuid;not null" json:"projectId"`
	BuyerID               uuid.UUID         `gorm:"type:uuid;not null" json:"buyerId"`
	SellerID              uuid.UUID         `gorm:"type:uuid;not null" json:"sellerId"`
	Amount                int               `gorm:"not null" json:"amount"`
	Status                TransactionStatus `gorm:"size:20;default:'pending'" json:"status"`
	MidtransOrderID       *string           `gorm:"size:255" json:"midtransOrderId,omitempty"`
	MidtransTransactionID *string           `gorm:"size:255" json:"midtransTransactionId,omitempty"`
	CreatedAt             time.Time         `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt             time.Time         `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Buyer   User    `gorm:"foreignKey:BuyerID" json:"buyer,omitempty"`
	Seller  User    `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type TransactionResponse struct {
	ID           uuid.UUID         `json:"id"`
	ProjectID    uuid.UUID         `json:"projectId"`
	ProjectTitle string            `json:"projectTitle"`
	BuyerName    string            `json:"buyerName"`
	Amount       int               `json:"amount"`
	Status       TransactionStatus `json:"status"`
	CreatedAt    time.Time         `json:"createdAt"`
}

func (t *Transaction) ToResponse() TransactionResponse {
	return TransactionResponse{
		ID:           t.ID,
		ProjectID:    t.ProjectID,
		ProjectTitle: t.Project.Title,
		BuyerName:    t.Buyer.Name,
		Amount:       t.Amount,
		Status:       t.Status,
		CreatedAt:    t.CreatedAt,
	}
}
