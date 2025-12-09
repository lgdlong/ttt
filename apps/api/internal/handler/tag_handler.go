package handler

import (
	"api/internal/dto"
	"api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	service service.TagService
}

func NewTagHandler(service service.TagService) *TagHandler {
	return &TagHandler{service: service}
}

// CreateTag godoc
// @Summary Create a new tag
// @Description Create a new tag (mod/admin only)
// @Tags Tags
// @Accept json
// @Produce json
// @Param tag body dto.CreateTagRequest true "Tag data"
// @Success 201 {object} dto.TagResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "Tag already exists"
// @Router /mod/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	tag, err := h.service.CreateTag(req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() != "" && err.Error()[:3] == "tag" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to create tag",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// GetTag godoc
// @Summary Get tag by ID
// @Description Get tag details by tag ID
// @Tags Tags
// @Accept json
// @Produce json
// @Param id path string true "Tag ID (UUID)"
// @Success 200 {object} dto.TagResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /mod/tags/{id} [get]
func (h *TagHandler) GetTag(c *gin.Context) {
	id := c.Param("id")

	tag, err := h.service.GetTagByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Tag not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// UpdateTag godoc
// @Summary Update tag
// @Description Update tag information (mod/admin only)
// @Tags Tags
// @Accept json
// @Produce json
// @Param id path string true "Tag ID (UUID)"
// @Param tag body dto.UpdateTagRequest true "Updated tag data"
// @Success 200 {object} dto.TagResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "Tag name already exists"
// @Router /mod/tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	tag, err := h.service.UpdateTag(id, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()
		if errMsg == "tag not found" || errMsg[:8] == "invalid " {
			statusCode = http.StatusNotFound
		} else if errMsg[:3] == "tag" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to update tag",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// DeleteTag godoc
// @Summary Delete tag
// @Description Delete a tag (mod/admin only)
// @Tags Tags
// @Accept json
// @Produce json
// @Param id path string true "Tag ID (UUID)"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /mod/tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteTag(id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Tag not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListTags godoc
// @Summary List tags
// @Description Get paginated list of tags with optional search
// @Tags Tags
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param query query string false "Search query"
// @Success 200 {object} dto.TagListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /mod/tags [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	var req dto.TagListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := h.service.ListTags(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list tags",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AddTagToVideo godoc
// @Summary Add tag to video
// @Description Add a tag to a video. Creates tag if tag_name provided and doesn't exist.
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Param request body dto.AddVideoTagRequest true "Tag info (provide tag_id OR tag_name)"
// @Success 200 {object} dto.TagResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /mod/videos/{id}/tags [post]
func (h *TagHandler) AddTagToVideo(c *gin.Context) {
	videoID := c.Param("id")

	var req dto.AddVideoTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	tag, err := h.service.AddTagToVideo(videoID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()
		if errMsg == "video not found" || errMsg == "tag not found" {
			statusCode = http.StatusNotFound
		} else if errMsg[:6] == "either" || errMsg[:7] == "invalid" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to add tag to video",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// RemoveTagFromVideo godoc
// @Summary Remove tag from video
// @Description Remove a tag from a video
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Param tag_id path string true "Tag ID (UUID)"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /mod/videos/{id}/tags/{tag_id} [delete]
func (h *TagHandler) RemoveTagFromVideo(c *gin.Context) {
	videoID := c.Param("id")
	tagID := c.Param("tag_id")

	if err := h.service.RemoveTagFromVideo(videoID, tagID); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Failed to remove tag",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetVideoTags godoc
// @Summary Get video tags
// @Description Get all tags for a video
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Success 200 {array} dto.TagResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /mod/videos/{id}/tags [get]
func (h *TagHandler) GetVideoTags(c *gin.Context) {
	videoID := c.Param("id")

	tags, err := h.service.GetVideoTags(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Failed to get video tags",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, tags)
}
