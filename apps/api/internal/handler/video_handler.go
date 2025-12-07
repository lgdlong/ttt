package handler

import (
	"api/internal/dto"
	"api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	service service.VideoService
}

func NewVideoHandler(service service.VideoService) *VideoHandler {
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
