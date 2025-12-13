package repository

import (
	"api/internal/domain"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type videoTranscriptReviewRepository struct {
	db *gorm.DB
}

func NewVideoTranscriptReviewRepository(db *gorm.DB) domain.VideoTranscriptReviewRepository {
	return &videoTranscriptReviewRepository{db: db}
}

// CreateReview creates a new review record
func (r *videoTranscriptReviewRepository) CreateReview(review *domain.VideoTranscriptReview) error {
	return r.db.Create(review).Error
}

// HasUserReviewedVideo checks if a user has already reviewed a specific video
func (r *videoTranscriptReviewRepository) HasUserReviewedVideo(videoID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.VideoTranscriptReview{}).
		Where("video_id = ? AND user_id = ?", videoID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetReviewCountByVideoID returns the total number of reviews for a video
func (r *videoTranscriptReviewRepository) GetReviewCountByVideoID(videoID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&domain.VideoTranscriptReview{}).
		Where("video_id = ?", videoID).
		Count(&count).Error

	return count, err
}

// GetReviewsByVideoID retrieves all reviews for a specific video
func (r *videoTranscriptReviewRepository) GetReviewsByVideoID(videoID uuid.UUID) ([]domain.VideoTranscriptReview, error) {
	var reviews []domain.VideoTranscriptReview
	err := r.db.Where("video_id = ?", videoID).
		Preload("User").
		Order("reviewed_at DESC").
		Find(&reviews).Error

	return reviews, err
}

// GetReviewsByUserID retrieves all reviews by a specific user
func (r *videoTranscriptReviewRepository) GetReviewsByUserID(userID uuid.UUID) ([]domain.VideoTranscriptReview, error) {
	var reviews []domain.VideoTranscriptReview
	err := r.db.Where("user_id = ?", userID).
		Preload("Video").
		Order("reviewed_at DESC").
		Find(&reviews).Error

	return reviews, err
}

// DeleteReview deletes a review by ID (admin only, rare use case)
func (r *videoTranscriptReviewRepository) DeleteReview(id uint) error {
	result := r.db.Delete(&domain.VideoTranscriptReview{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}
