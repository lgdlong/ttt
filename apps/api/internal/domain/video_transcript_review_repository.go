package domain

import (
	"github.com/google/uuid"
)

type VideoTranscriptReviewRepository interface {
	// Create a new review
	CreateReview(review *VideoTranscriptReview) error

	// Check if user has already reviewed this video
	HasUserReviewedVideo(videoID, userID uuid.UUID) (bool, error)

	// Get review count for a specific video
	GetReviewCountByVideoID(videoID uuid.UUID) (int64, error)

	// Get all reviews for a video
	GetReviewsByVideoID(videoID uuid.UUID) ([]VideoTranscriptReview, error)

	// Get all reviews by a specific user
	GetReviewsByUserID(userID uuid.UUID) ([]VideoTranscriptReview, error)

	// Delete a review (admin only, rare use case)
	DeleteReview(id uint) error
}
