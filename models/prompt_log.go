package models

import (
	"time"

	"github.com/google/uuid"
)

// PromptLog stores a scored prompt from a student's GenAI interaction.
// CriticalityScore is a regression value in [0, 1]: 0 = lazy, 1 = critical.
type PromptLog struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID        uuid.UUID `gorm:"type:uuid;not null;index" json:"session_id"`
	Session          Session   `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"-"`
	PromptText       string    `gorm:"type:text;not null" json:"prompt_text"`
	CriticalityScore float64   `gorm:"type:decimal(3,2);not null" json:"criticality_score"`
	CreatedAt        time.Time `json:"created_at"`
}
