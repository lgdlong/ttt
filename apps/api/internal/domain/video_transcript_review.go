package domain

import (
	"time"

	"github.com/google/uuid"
)

// VideoTranscriptReview tracks moderator reviews for video transcripts
// Used for KPI tracking and video verification workflow
type VideoTranscriptReview struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VideoID    uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ReviewedAt time.Time `gorm:"type:timestamptz;default:now();not null;index" json:"reviewed_at"`

	// Relationships
	Video *Video `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE" json:"video,omitempty"`
	User  *User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

func (VideoTranscriptReview) TableName() string {
	return "video_transcript_reviews"
}
