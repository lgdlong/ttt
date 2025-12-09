package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/repository"
	"fmt"
	"time"

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

type videoTranscriptReviewService struct {
	reviewRepo repository.VideoTranscriptReviewRepository
	videoRepo  repository.VideoRepository
	userRepo   repository.UserRepository
}

func NewVideoTranscriptReviewService(
	reviewRepo repository.VideoTranscriptReviewRepository,
	videoRepo repository.VideoRepository,
	userRepo repository.UserRepository,
) VideoTranscriptReviewService {
	return &videoTranscriptReviewService{
		reviewRepo: reviewRepo,
		videoRepo:  videoRepo,
		userRepo:   userRepo,
	}
}

// SubmitReview handles the review submission workflow
func (s *videoTranscriptReviewService) SubmitReview(
	videoID, userID uuid.UUID,
	req dto.SubmitReviewRequest,
) (*dto.VideoTranscriptReviewResponse, error) {
	// 1. Cho phép user review nhiều lần (bỏ kiểm tra đã review)

	// 2. Get video details (removed unused variable)
	if _, err := s.videoRepo.GetVideoByID(videoID); err != nil {
		return nil, fmt.Errorf("video not found: %w", err)
	}

	// 3. Create review record
	review := &domain.VideoTranscriptReview{
		VideoID:    videoID,
		UserID:     userID,
		ReviewedAt: time.Now(),
	}

	if err := s.reviewRepo.CreateReview(review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// 4. (Removed) Award points to user - not required

	// 5. Get total review count for this video
	totalReviews, err := s.reviewRepo.GetReviewCountByVideoID(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get review count: %w", err)
	}

	// 6. (Removed) Update video status logic - not required
	statusMessage := fmt.Sprintf("Review submitted. Total reviews: %d", totalReviews)

	// 7. Build response
	response := &dto.VideoTranscriptReviewResponse{
		ID:           review.ID,
		VideoID:      videoID.String(),
		UserID:       userID.String(),
		ReviewedAt:   review.ReviewedAt.Format(time.RFC3339),
		TotalReviews: int(totalReviews),
		Message:      statusMessage,
	}

	return response, nil
}

// GetVideoReviewStats returns review count for a video
func (s *videoTranscriptReviewService) GetVideoReviewStats(videoID uuid.UUID) (int64, error) {
	return s.reviewRepo.GetReviewCountByVideoID(videoID)
}

// HasUserReviewedVideo checks if a user has already reviewed a video
func (s *videoTranscriptReviewService) HasUserReviewedVideo(videoID, userID uuid.UUID) (bool, error) {
	return s.reviewRepo.HasUserReviewedVideo(videoID, userID)
}
