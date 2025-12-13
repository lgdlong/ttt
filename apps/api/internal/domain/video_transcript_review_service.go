package domain

import (
	"api/internal/dto"

	"github.com/google/uuid"
)

type VideoTranscriptReviewService interface {
	// Submit a review for a video
	SubmitReview(videoID, userID uuid.UUID, req dto.SubmitReviewRequest) (*dto.VideoTranscriptReviewResponse, error)

	// Get review statistics for a video
	GetVideoReviewStats(videoID uuid.UUID) (int64, error)

	// Check if user has reviewed a video
	HasUserReviewedVideo(videoID, userID uuid.UUID) (bool, error)
}
