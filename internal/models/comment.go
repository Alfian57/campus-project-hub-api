package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null" json:"projectId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// Relationships
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type CommentResponse struct {
	ID        uuid.UUID    `json:"id"`
	Content   string       `json:"content"`
	User      UserResponse `json:"user"`
	CreatedAt time.Time    `json:"createdAt"`
}

func (c *Comment) ToResponse() CommentResponse {
	return CommentResponse{
		ID:        c.ID,
		Content:   c.Content,
		User:      c.User.ToResponse(),
		CreatedAt: c.CreatedAt,
	}
}
