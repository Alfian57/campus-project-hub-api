package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string
type UserStatus string

const (
	RoleUser      UserRole = "user"
	RoleAdmin     UserRole = "admin"
	RoleModerator UserRole = "moderator"
)

const (
	StatusActive  UserStatus = "active"
	StatusBlocked UserStatus = "blocked"
)

type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email         string     `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash  *string    `gorm:"size:255" json:"-"`
	Name          string     `gorm:"not null;size:255" json:"name"`
	AvatarURL     *string    `gorm:"type:text" json:"avatarUrl"`
	University    *string    `gorm:"size:255" json:"university"`
	Major         *string    `gorm:"size:255" json:"major"`
	Bio           *string    `gorm:"type:text" json:"bio"`
	Phone         *string    `gorm:"size:20" json:"phone"`
	Role          UserRole   `gorm:"size:20;default:'user'" json:"role"`
	Status        UserStatus `gorm:"size:20;default:'active'" json:"status"`
	TotalExp      int        `gorm:"default:0" json:"totalExp"`
	OAuthProvider *string    `gorm:"column:oauth_provider;size:20" json:"oauthProvider,omitempty"`
	OAuthID       *string    `gorm:"column:oauth_id;size:255" json:"-"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Projects      []Project `gorm:"foreignKey:UserID" json:"projects,omitempty"`
	Articles      []Article `gorm:"foreignKey:UserID" json:"articles,omitempty"`
	Comments      []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	LikedProjects []Project `gorm:"many2many:project_likes" json:"likedProjects,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserResponse is the safe response without sensitive data
type UserResponse struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	AvatarURL  *string    `json:"avatarUrl"`
	University *string    `json:"university"`
	Major      *string    `json:"major"`
	Bio        *string    `json:"bio"`
	Phone      *string    `json:"phone"`
	Role       UserRole   `json:"role"`
	Status     UserStatus `json:"status"`
	TotalExp   int        `json:"totalExp"`
	Level      int        `json:"level"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		AvatarURL:  u.AvatarURL,
		University: u.University,
		Major:      u.Major,
		Bio:        u.Bio,
		Phone:      u.Phone,
		Role:       u.Role,
		Status:     u.Status,
		TotalExp:   u.TotalExp,
		Level:      GetLevelFromExp(u.TotalExp),
		CreatedAt:  u.CreatedAt,
	}
}

// Gamification helpers
func GetLevelFromExp(totalExp int) int {
	level := 1
	for level < 100 {
		required := 100 * level * level
		if totalExp < required {
			break
		}
		level++
	}
	return level
}
