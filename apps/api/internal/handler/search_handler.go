package handler

import (
	"api/internal/dto"
	"api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	service service.VideoService
}

func NewSearchHandler(service service.VideoService) *SearchHandler {
	return &SearchHandler{service: service}
}

// SearchTranscript godoc
// @Summary Search transcripts by text
// @Description Full-text search across all transcript segments using PostgreSQL FTS
// @Tags Search
// @Accept json
// @Produce json
// @Param q query string true "Search query (supports PostgreSQL websearch syntax)" minlength(2)
// @Param limit query int false "Number of results" default(20) minimum(1) maximum(50)
// @Success 200 {object} dto.TranscriptSearchResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /search/transcript [get]
func (h *SearchHandler) SearchTranscript(c *gin.Context) {
	var req dto.TranscriptSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := h.service.SearchTranscripts(req)
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

// SearchTags godoc
// @Summary Search tags by semantic similarity
// @Description Semantic search using vector embeddings and cosine similarity
// @Tags Search
// @Accept json
// @Produce json
// @Param q query string true "Search query" minlength(2)
// @Param limit query int false "Number of results" default(5) minimum(1) maximum(10)
// @Success 200 {object} dto.TagSearchResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /search/tags [get]
func (h *SearchHandler) SearchTags(c *gin.Context) {
	var req dto.TagSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := h.service.SearchTags(req)
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
