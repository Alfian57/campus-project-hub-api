package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TargetType string

const (
	TargetTypeUser    TargetType = "user"
	TargetTypeProject TargetType = "project"
	TargetTypeComment TargetType = "comment"
)

type BlockRecord struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TargetType TargetType `gorm:"size:20;not null" json:"targetType"`
	TargetID   uuid.UUID  `gorm:"type:uuid;not null" json:"targetId"`
	TargetName string     `gorm:"size:255" json:"targetName"`
	Reason     *string    `gorm:"type:text" json:"reason"`
	BlockedBy  uuid.UUID  `gorm:"type:uuid;not null" json:"blockedBy"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"blockedAt"`

	// Relationships
	Blocker User `gorm:"foreignKey:BlockedBy" json:"blocker,omitempty"`
}

func (b *BlockRecord) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type BlockRecordResponse struct {
	ID            uuid.UUID  `json:"id"`
	TargetType    TargetType `json:"targetType"`
	TargetID      uuid.UUID  `json:"targetId"`
	TargetName    string     `json:"targetName"`
	Reason        *string    `json:"reason"`
	BlockedBy     uuid.UUID  `json:"blockedBy"`
	BlockedByName string     `json:"blockedByName"`
	BlockedAt     time.Time  `json:"blockedAt"`
}

func (b *BlockRecord) ToResponse() BlockRecordResponse {
	return BlockRecordResponse{
		ID:            b.ID,
		TargetType:    b.TargetType,
		TargetID:      b.TargetID,
		TargetName:    b.TargetName,
		Reason:        b.Reason,
		BlockedBy:     b.BlockedBy,
		BlockedByName: b.Blocker.Name,
		BlockedAt:     b.CreatedAt,
	}
}
