package repository

import (
	"api/internal/domain"
	"time"

	"gorm.io/gorm"
)

type statsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) domain.StatsRepository {
	return &statsRepository{db: db}
}

// GetTotalUsers returns total number of users
func (r *statsRepository) GetTotalUsers() (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Count(&count).Error
	return count, err
}

// GetActiveUsers returns number of non-banned users
// GetActiveUsers returns number of active users
func (r *statsRepository) GetActiveUsers() (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// GetTotalVideos returns total number of videos (excluding soft-deleted)
func (r *statsRepository) GetTotalVideos() (int64, error) {
	var count int64
	err := r.db.Model(&domain.Video{}).Count(&count).Error
	return count, err
}

// GetTotalTags returns total number of canonical tags
func (r *statsRepository) GetTotalTags() (int64, error) {
	var count int64
	err := r.db.Model(&domain.CanonicalTag{}).Count(&count).Error
	return count, err
}

// GetVideosWithTranscript returns number of videos that have transcript segments
func (r *statsRepository) GetVideosWithTranscript() (int64, error) {
	var count int64
	err := r.db.Model(&domain.Video{}).
		Joins("JOIN transcript_segments ON videos.id = transcript_segments.video_id").
		Distinct("videos.id").
		Count(&count).Error
	return count, err
}

// GetVideosAddedToday returns number of videos added today
func (r *statsRepository) GetVideosAddedToday() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Model(&domain.Video{}).
		Where("created_at >= ?", today).
		Count(&count).Error
	return count, err
}
