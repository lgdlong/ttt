package handler

import (
	"api/internal/domain"
	"api/internal/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VideoTranscriptReviewHandler struct {
	service domain.VideoTranscriptReviewService
}

func NewVideoTranscriptReviewHandler(service domain.VideoTranscriptReviewService) *VideoTranscriptReviewHandler {
	return &VideoTranscriptReviewHandler{service: service}
}

// SubmitReview godoc
// @Summary Submit transcript review for a video
// @Description Moderator submits a review after verifying video transcript. Awards points and updates video status if threshold met.
// @Tags Video Reviews
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Param request body dto.SubmitReviewRequest false "Optional review notes"
// @Success 200 {object} dto.VideoTranscriptReviewResponse
// @Failure 400 {object} dto.ErrorResponse "Invalid request or user already reviewed"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Video not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /videos/{id}/reviews [post]
// @Security BearerAuth
func (h *VideoTranscriptReviewHandler) SubmitReview(c *gin.Context) {
	// 1. Get video ID from path
	videoIDStr := c.Param("id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid video ID",
			Message: "Video ID must be a valid UUID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 2. Get user ID from JWT context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var userID uuid.UUID
	switch v := userIDInterface.(type) {
	case uuid.UUID:
		userID = v
	case string:
		var err error
		userID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Internal error",
				Message: "Failed to parse user ID",
				Code:    http.StatusInternalServerError,
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal error",
			Message: "Invalid user ID format",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 3. Bind request body (optional notes)
	var req dto.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body, just use default values
		req = dto.SubmitReviewRequest{}
	}

	// 4. Call service to submit review
	response, err := h.service.SubmitReview(videoID, userID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "Failed to submit review"

		// Handle specific errors
		if err.Error() == "user has already reviewed this video" {
			statusCode = http.StatusBadRequest
			errorMsg = "You have already reviewed this video"
		} else if err.Error() == "video not found: record not found" {
			statusCode = http.StatusNotFound
			errorMsg = "Video not found"
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   errorMsg,
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	// 5. Return success response
	c.JSON(http.StatusOK, response)
}

// GetVideoReviewStats godoc
// @Summary Get review statistics for a video
// @Description Get the number of reviews for a specific video
// @Tags Video Reviews
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Success 200 {object} map[string]interface{} "Returns {video_id, review_count}"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /videos/{id}/reviews/stats [get]
func (h *VideoTranscriptReviewHandler) GetVideoReviewStats(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid video ID",
			Message: "Video ID must be a valid UUID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	count, err := h.service.GetVideoReviewStats(videoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get review stats",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"video_id":     videoIDStr,
		"review_count": count,
	})
}

// CheckUserReviewStatus godoc
// @Summary Check if user has reviewed a video
// @Description Check whether the authenticated user has already reviewed a specific video
// @Tags Video Reviews
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Success 200 {object} map[string]interface{} "Returns {video_id, has_reviewed}"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /videos/{id}/reviews/status [get]
// @Security BearerAuth
func (h *VideoTranscriptReviewHandler) CheckUserReviewStatus(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid video ID",
			Message: "Video ID must be a valid UUID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userIDStr, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal error",
			Message: "Invalid user ID format",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal error",
			Message: "Failed to parse user ID",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	hasReviewed, err := h.service.HasUserReviewedVideo(videoID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to check review status",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"video_id":     videoIDStr,
		"has_reviewed": hasReviewed,
	})
}
