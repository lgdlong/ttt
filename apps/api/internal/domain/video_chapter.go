package domain

import (
	"time"

	"github.com/google/uuid"
)

type VideoChapter struct {
	ID           uint      `gorm:"primaryKey"`
	VideoID      uuid.UUID `gorm:"type:uuid;index;not null"`
	Title        string    `gorm:"type:text;not null"`
	Content      string    `gorm:"type:text;not null"`
	StartTime    int       `gorm:"not null;default:0"`
	ChapterOrder int       `gorm:"not null;default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (VideoChapter) TableName() string {
	return "video_chapters"
}
