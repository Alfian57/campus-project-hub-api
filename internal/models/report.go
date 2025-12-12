package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusResolved ReportStatus = "resolved"
	ReportStatusRejected ReportStatus = "rejected"
)

type Report struct {
	ID         uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReporterID uuid.UUID    `gorm:"type:uuid;not null" json:"reporterId"`
	TargetType TargetType   `gorm:"size:20;not null" json:"targetType"`
	TargetID   uuid.UUID    `gorm:"type:uuid;not null" json:"targetId"`
	TargetName string       `gorm:"size:255" json:"targetName"`
	Reason     *string      `gorm:"type:text" json:"reason"`
	Status     ReportStatus `gorm:"size:20;default:'pending'" json:"status"`
	ResolvedBy *uuid.UUID   `gorm:"type:uuid" json:"resolvedBy,omitempty"`
	ResolvedAt *time.Time   `json:"resolvedAt,omitempty"`
	CreatedAt  time.Time    `gorm:"autoCreateTime" json:"createdAt"`

	// Relationships
	Reporter User  `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	Resolver *User `gorm:"foreignKey:ResolvedBy" json:"resolver,omitempty"`
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type ReportResponse struct {
	ID           uuid.UUID    `json:"id"`
	ReporterName string       `json:"reporterName"`
	TargetType   TargetType   `json:"targetType"`
	TargetID     uuid.UUID    `json:"targetId"`
	TargetName   string       `json:"targetName"`
	Reason       *string      `json:"reason"`
	Status       ReportStatus `json:"status"`
	ResolvedBy   *string      `json:"resolvedBy,omitempty"`
	ResolvedAt   *time.Time   `json:"resolvedAt,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
}

func (r *Report) ToResponse() ReportResponse {
	var resolvedByName *string
	if r.Resolver != nil {
		resolvedByName = &r.Resolver.Name
	}

	return ReportResponse{
		ID:           r.ID,
		ReporterName: r.Reporter.Name,
		TargetType:   r.TargetType,
		TargetID:     r.TargetID,
		TargetName:   r.TargetName,
		Reason:       r.Reason,
		Status:       r.Status,
		ResolvedBy:   resolvedByName,
		ResolvedAt:   r.ResolvedAt,
		CreatedAt:    r.CreatedAt,
	}
}
