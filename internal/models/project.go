package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ProjectType string
type ProjectStatus string

const (
	ProjectTypeFree ProjectType = "free"
	ProjectTypePaid ProjectType = "paid"
)

const (
	ProjectStatusPublished ProjectStatus = "published"
	ProjectStatusDraft     ProjectStatus = "draft"
	ProjectStatusBlocked   ProjectStatus = "blocked"
)

type Project struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Title        string         `gorm:"not null;size:255" json:"title"`
	Description  *string        `gorm:"type:text" json:"description"`
	ThumbnailURL *string        `gorm:"type:text" json:"thumbnailUrl"`
	TechStack    pq.StringArray `gorm:"type:text[]" json:"techStack"`
	GithubURL    *string        `gorm:"type:text" json:"githubUrl"`
	DemoURL      *string        `gorm:"type:text" json:"demoUrl"`
	Type         ProjectType    `gorm:"size:10;default:'free'" json:"type"`
	Price        int            `gorm:"default:0" json:"price"`
	Status       ProjectStatus  `gorm:"size:20;default:'published'" json:"status"`
	Views        int            `gorm:"default:0" json:"views"`
	Likes        int            `gorm:"default:0" json:"likes"`
	CategoryID   *uuid.UUID     `gorm:"type:uuid" json:"categoryId"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	User     User           `gorm:"foreignKey:UserID" json:"author,omitempty"`
	Category *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images   []ProjectImage `gorm:"foreignKey:ProjectID" json:"images,omitempty"`
	Comments []Comment      `gorm:"foreignKey:ProjectID" json:"comments,omitempty"`
	LikedBy  []User         `gorm:"many2many:project_likes" json:"likedBy,omitempty"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type ProjectImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null" json:"projectId"`
	ImageURL  string    `gorm:"type:text;not null" json:"imageUrl"`
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
}

func (pi *ProjectImage) BeforeCreate(tx *gorm.DB) error {
	if pi.ID == uuid.Nil {
		pi.ID = uuid.New()
	}
	return nil
}

type ProjectLike struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	ProjectID uuid.UUID `gorm:"type:uuid;primaryKey" json:"projectId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// ProjectResponse for API response
type ProjectResponse struct {
	ID           uuid.UUID     `json:"id"`
	Title        string        `json:"title"`
	Description  *string       `json:"description"`
	ThumbnailURL *string       `json:"thumbnailUrl"`
	Images       []string      `json:"images"`
	TechStack    []string      `json:"techStack"`
	Links        ProjectLinks  `json:"links"`
	Stats        ProjectStats  `json:"stats"`
	Type         ProjectType   `json:"type"`
	Price        int           `json:"price,omitempty"`
	Status       ProjectStatus `json:"status"`
	Author       UserResponse  `json:"author"`
	CategoryID   *uuid.UUID    `json:"categoryId"`
	CreatedAt    time.Time     `json:"createdAt"`
}

type ProjectLinks struct {
	Github string `json:"github"`
	Demo   string `json:"demo"`
}

type ProjectStats struct {
	Views        int `json:"views"`
	Likes        int `json:"likes"`
	CommentCount int `json:"commentCount"`
}

func (p *Project) ToResponse(commentCount int) ProjectResponse {
	images := make([]string, len(p.Images))
	for i, img := range p.Images {
		images[i] = img.ImageURL
	}

	var githubURL, demoURL string
	if p.GithubURL != nil {
		githubURL = *p.GithubURL
	}
	if p.DemoURL != nil {
		demoURL = *p.DemoURL
	}

	return ProjectResponse{
		ID:           p.ID,
		Title:        p.Title,
		Description:  p.Description,
		ThumbnailURL: p.ThumbnailURL,
		Images:       images,
		TechStack:    p.TechStack,
		Links: ProjectLinks{
			Github: githubURL,
			Demo:   demoURL,
		},
		Stats: ProjectStats{
			Views:        p.Views,
			Likes:        p.Likes,
			CommentCount: commentCount,
		},
		Type:       p.Type,
		Price:      p.Price,
		Status:     p.Status,
		Author:     p.User.ToResponse(),
		CategoryID: p.CategoryID,
		CreatedAt:  p.CreatedAt,
	}
}
