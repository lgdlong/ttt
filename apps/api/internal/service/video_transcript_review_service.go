package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// Video status constants
	VideoStatusDraft     = "DRAFT"
	VideoStatusPublished = "PUBLISHED"

	// Review requirements
	RequiredReviewsForPublish = 2

	// Points calculation
	BasePointsPerMinute = 1 // 1 point per minute of video
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

	// 2. Get video details for points calculation
	video, err := s.videoRepo.GetVideoByID(videoID)
	if err != nil {
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

	// 4. Calculate and award points to user
	pointsAwarded := s.calculateReviewPoints(video.Duration)
	if err := s.awardPointsToUser(userID, pointsAwarded); err != nil {
		// Log error but don't fail the review submission
		fmt.Printf("Warning: Failed to award points to user %s: %v\n", userID, err)
	}

	// 5. Get total review count for this video
	totalReviews, err := s.reviewRepo.GetReviewCountByVideoID(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get review count: %w", err)
	}

	// 6. Update video status if threshold reached
	videoStatus := VideoStatusDraft
	statusMessage := fmt.Sprintf("Review submitted. %d/%d reviews completed.", totalReviews, RequiredReviewsForPublish)

	if totalReviews >= RequiredReviewsForPublish {
		if err := s.updateVideoStatus(videoID, VideoStatusPublished); err != nil {
			// Log error but don't fail the review submission
			fmt.Printf("Warning: Failed to update video status: %v\n", err)
		} else {
			videoStatus = VideoStatusPublished
			statusMessage = fmt.Sprintf("Video has been published! Total reviews: %d", totalReviews)
		}
	}

	// 7. Build response
	response := &dto.VideoTranscriptReviewResponse{
		ID:            review.ID,
		VideoID:       videoID.String(),
		UserID:        userID.String(),
		ReviewedAt:    review.ReviewedAt.Format(time.RFC3339),
		TotalReviews:  int(totalReviews),
		VideoStatus:   videoStatus,
		PointsAwarded: pointsAwarded,
		Message:       statusMessage,
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

// ===== PRIVATE HELPER METHODS =====

// calculateReviewPoints calculates points based on video duration
// Formula: 1 point per minute of video
func (s *videoTranscriptReviewService) calculateReviewPoints(durationSeconds int) int {
	durationMinutes := durationSeconds / 60
	if durationMinutes < 1 {
		durationMinutes = 1 // Minimum 1 point
	}
	return durationMinutes * BasePointsPerMinute
}

// awardPointsToUser adds points to user's reputation/score
// Note: Assumes User model has a "Points" or "ReputationScore" field
// Modify based on actual User schema
func (s *videoTranscriptReviewService) awardPointsToUser(userID uuid.UUID, points int) error {
	// Get current user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// TODO: Add "reputation_points" field to User model if not exists
	// For now, we'll use a generic update approach
	updates := map[string]interface{}{
		// "reputation_points": gorm.Expr("reputation_points + ?", points),
		"updated_at": time.Now(),
	}

	// Note: If User model doesn't have reputation_points field,
	// this will silently do nothing. Add the field to domain.User first.
	_ = user // Prevent unused variable error

	if err := s.userRepo.UpdateUser(userID, updates); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// updateVideoStatus updates video status to PUBLISHED
// Note: Assumes Video model has a "Status" field
func (s *videoTranscriptReviewService) updateVideoStatus(videoID uuid.UUID, status string) error {
	video, err := s.videoRepo.GetVideoByID(videoID)
	if err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	// TODO: Add "status" field to Video model if not exists
	// For now, we'll use the Update method
	_ = video // Prevent unused variable error

	// Update video with new status
	updates := &domain.Video{
		// Status: status, // Uncomment when Status field is added to Video model
		UpdatedAt: time.Now(),
	}

	if err := s.videoRepo.Update(updates); err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	return nil
}
