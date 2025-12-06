package domain

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	YoutubeID    string    `gorm:"type:varchar(20);uniqueIndex;not null"`
	Title        string    `gorm:"type:text;not null"`
	PublishedAt  time.Time `gorm:"type:date;index"`
	Duration     int       `gorm:"not null"` // Seconds
	ViewCount    int       `gorm:"default:0"`
	ThumbnailURL string    `gorm:"type:varchar(500)"`

	// Relationship 1-N: Subtitles
	Segments []TranscriptSegment `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE;"`

	// Relationship N-N: Tags (GORM tự xử lý bảng trung gian video_tags)
	Tags []Tag `gorm:"many2many:video_tags;"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (Video) TableName() string {
	return "videos"
}
