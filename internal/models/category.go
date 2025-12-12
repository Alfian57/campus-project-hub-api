package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null;size:100" json:"name"`
	Slug        string    `gorm:"uniqueIndex;not null;size:100" json:"slug"`
	Description *string   `gorm:"type:text" json:"description"`
	Color       *string   `gorm:"size:20" json:"color"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// Computed field (not stored in DB)
	ProjectCount int `gorm:"-" json:"projectCount"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
