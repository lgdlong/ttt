package handler

import (
	"api/internal/domain"
	"api/internal/dto"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	service domain.VideoService
}

func NewVideoHandler(service domain.VideoService) *VideoHandler {
	return &VideoHandler{service: service}
}

// GetVideoList godoc
// @Summary List videos with pagination
// @Description Get a paginated list of videos with optional filtering and sorting
// @Tags Videos
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(50)
// @Param sort query string false "Sort order" Enums(newest, popular, views)
// @Param tag_id query string false "Filter by tag ID (UUID)"
// @Success 200 {object} dto.VideoListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /videos [get]
func (h *VideoHandler) GetVideoList(c *gin.Context) {
	var req dto.ListVideoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := h.service.GetVideoList(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to fetch videos",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetVideoDetail godoc
// @Summary Get video details
// @Description Get detailed information about a specific video including tags
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Success 200 {object} dto.VideoDetailResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /videos/{id} [get]
func (h *VideoHandler) GetVideoDetail(c *gin.Context) {
	id := c.Param("id")

	response, err := h.service.GetVideoDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Video not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetVideoTranscript godoc
// @Summary Get video transcript
// @Description Get all transcript segments for a specific video
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Success 200 {object} dto.TranscriptResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /videos/{id}/transcript [get]
func (h *VideoHandler) GetVideoTranscript(c *gin.Context) {
	id := c.Param("id")

	response, err := h.service.GetVideoTranscript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Transcript not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSegment godoc
// @Summary Update transcript segment
// @Description Update text content of a single transcript segment
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path int true "Segment ID"
// @Param request body dto.UpdateSegmentRequest true "Updated segment data"
// @Success 200 {object} dto.SegmentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /transcript-segments/{id} [patch]
func (h *VideoHandler) UpdateSegment(c *gin.Context) {
	var segmentID uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &segmentID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid segment ID",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.UpdateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := h.service.UpdateSegment(segmentID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Failed to update segment",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateSegment godoc
// @Summary Create new transcript segment
// @Description Add a new transcript segment to a video
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Param request body dto.CreateSegmentRequest true "Segment data"
// @Success 201 {object} dto.SegmentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /mod/videos/{id}/transcript/segments [post]
func (h *VideoHandler) CreateSegment(c *gin.Context) {
	videoID := c.Param("id")

	var req dto.CreateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := h.service.CreateSegment(videoID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Failed to create segment",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetModVideoList godoc
// @Summary List videos for mod dashboard
// @Description Get paginated videos for mod/admin dashboard with tag information
// @Tags Videos
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param page_size query int false "Items per page" default(10) minimum(1) maximum(50)
// @Param q query string false "Search by title"
// @Param tag_ids query string false "Filter by tag IDs (comma-separated)"
// @Param has_transcript query string false "Filter by transcript status" Enums(all,true,false)
// @Success 200 {object} dto.ModVideoListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /mod/videos [get]
func (h *VideoHandler) GetModVideoList(c *gin.Context) {
	// Parse pagination params
	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if _, err := fmt.Sscanf(ps, "%d", &pageSize); err != nil || pageSize < 1 || pageSize > 50 {
			pageSize = 10
		}
	}

	searchQuery := c.Query("q")
	tagIDsStr := c.Query("tag_ids")
	hasTranscriptStr := c.Query("has_transcript") // "all", "true", or "false"

	// Get videos from service
	videos, total, err := h.service.GetModVideoList(page, pageSize, searchQuery, tagIDsStr, hasTranscriptStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to fetch videos",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, dto.ModVideoListResponse{
		Videos:   videos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// CreateVideo godoc
// @Summary Create video from YouTube
// @Description Create a new video by fetching metadata from YouTube (mod/admin only)
// @Tags Videos
// @Accept json
// @Produce json
// @Param request body dto.CreateVideoRequest true "YouTube video ID"
// @Success 201 {object} dto.VideoCreateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "Video already exists"
// @Router /mod/videos [post]
func (h *VideoHandler) CreateVideo(c *gin.Context) {
	var req dto.CreateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := h.service.CreateVideo(req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()
		if len(errMsg) > 5 && errMsg[:5] == "video" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to create video",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// PreviewVideo godoc
// @Summary Preview YouTube video metadata
// @Description Fetch YouTube metadata without saving (mod/admin only)
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "YouTube video ID"
// @Success 200 {object} dto.VideoCreateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /mod/videos/preview/{id} [get]
func (h *VideoHandler) PreviewVideo(c *gin.Context) {
	youtubeID := c.Param("id")
	if youtubeID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Missing YouTube ID",
			Message: "YouTube video ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Fetch metadata from YouTube without creating video
	response, err := h.service.PreviewYouTubeVideo(youtubeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Failed to fetch video metadata",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteVideo godoc
// @Summary Delete video (soft delete)
// @Description Soft delete a video (mod/admin only)
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /mod/videos/{id} [delete]
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteVideo(id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Video not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchVideos godoc
// @Summary Search videos
// @Description Search videos by title (mod/admin only)
// @Tags Videos
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.VideoListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /mod/videos/search [get]
func (h *VideoHandler) SearchVideos(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Missing search query",
			Message: "Query parameter 'q' is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := parsePositiveInt(p); err == nil {
			page = parsed
		}
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := parsePositiveInt(l); err == nil {
			limit = parsed
		}
	}

	response, err := h.service.SearchVideos(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to parse positive int
func parsePositiveInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil || i < 1 {
		return 1, fmt.Errorf("invalid positive integer")
	}
	return i, nil
}
