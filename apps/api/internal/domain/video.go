package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Video struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	YoutubeID     string    `gorm:"type:varchar(20);uniqueIndex;not null"`
	Title         string    `gorm:"type:text;not null"`
	PublishedAt   time.Time `gorm:"type:date;index"`
	Duration      int       `gorm:"not null"` // Seconds
	ViewCount     int       `gorm:"default:0"`
	ThumbnailURL  string    `gorm:"type:varchar(500)"`
	HasTranscript bool      `gorm:"default:false;not null"` // TRUE nếu có ít nhất 1 segment
	Summary       string    `gorm:"type:text"`              // New: LLM Summary

	// Relationship 1-N: Subtitles
	Segments []TranscriptSegment `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE;"`

	// Relationship 1-N: Chapters (Semantic)
	Chapters []VideoChapter `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE;"`

	// Relationship N-N: CanonicalTags (GORM tự xử lý bảng trung gian video_canonical_tags)
	CanonicalTags []CanonicalTag `gorm:"many2many:video_canonical_tags;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // <--- SOFT DELETE
}

func (Video) TableName() string {
	return "videos"
}
