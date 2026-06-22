package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Report stores a generated analysis report for a session.
type Report struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SessionID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"session_id"`
	Session       Session   `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"-"`
	ReportContent string    `gorm:"type:text;not null" json:"report_content"`
	OverallScore  float64   `gorm:"type:decimal(3,2);not null" json:"overall_score"`
	CreatedAt     time.Time `json:"created_at"`
}

// BeforeCreate generates a new UUID if one is not already set.
func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
