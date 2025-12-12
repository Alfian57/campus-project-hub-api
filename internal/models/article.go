package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArticleStatus string

const (
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusBlocked   ArticleStatus = "blocked"
)

type Article struct {
	ID           uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID     `gorm:"type:uuid;not null" json:"userId"`
	Title        string        `gorm:"not null;size:255" json:"title"`
	Excerpt      *string       `gorm:"type:text" json:"excerpt"`
	Content      *string       `gorm:"type:text" json:"content"`
	ThumbnailURL *string       `gorm:"type:text" json:"thumbnailUrl"`
	Category     *string       `gorm:"size:100" json:"category"`
	ReadingTime  int           `gorm:"default:0" json:"readingTime"`
	Status       ArticleStatus `gorm:"size:20;default:'published'" json:"status"`
	Views        int           `gorm:"default:0" json:"views"`
	PublishedAt  *time.Time    `json:"publishedAt"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"author,omitempty"`
}

func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type ArticleResponse struct {
	ID           uuid.UUID     `json:"id"`
	Title        string        `json:"title"`
	Excerpt      *string       `json:"excerpt"`
	Content      *string       `json:"content"`
	ThumbnailURL *string       `json:"thumbnailUrl"`
	Category     *string       `json:"category"`
	ReadingTime  int           `json:"readingTime"`
	Status       ArticleStatus `json:"status"`
	Views        int           `json:"views"`
	PublishedAt  *time.Time    `json:"publishedAt"`
	Author       UserResponse  `json:"author"`
	CreatedAt    time.Time     `json:"createdAt"`
}

func (a *Article) ToResponse() ArticleResponse {
	return ArticleResponse{
		ID:           a.ID,
		Title:        a.Title,
		Excerpt:      a.Excerpt,
		Content:      a.Content,
		ThumbnailURL: a.ThumbnailURL,
		Category:     a.Category,
		ReadingTime:  a.ReadingTime,
		Status:       a.Status,
		Views:        a.Views,
		PublishedAt:  a.PublishedAt,
		Author:       a.User.ToResponse(),
		CreatedAt:    a.CreatedAt,
	}
}
