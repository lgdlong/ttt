package handler

import (
	"api/internal/dto"
	"api/internal/service"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	service   service.TagService
	serviceV2 service.TagServiceV2
}

func NewTagHandler(service service.TagService, serviceV2 service.TagServiceV2) *TagHandler {
	return &TagHandler{
		service:   service,
		serviceV2: serviceV2,
	}
}

/*
// ============================================================
// Legacy Tag Handlers - REMOVED (use Tag V2 API)
// ============================================================

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
	startTime := time.Now()
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[ERROR] Invalid JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	fmt.Printf("[CREATE_TAG] Received: '%s'\n", req.Name)
	tag, err := h.service.CreateTag(c.Request.Context(), req)
	if err != nil {
		// Check if it's a semantic duplicate error
		var semanticErr *service.SemanticDuplicateError
		if errors.As(err, &semanticErr) {
			// Calculate similarity percentage
			similarity := (1.0 - semanticErr.Distance/2.0) * 100

			// Log semantic duplicate
			fmt.Printf("[SEMANTIC_DUP] '%s' matches '%s' (distance: %.4f, similarity: %.1f%%)\n",
				req.Name, semanticErr.ExistingTag.Name, semanticErr.Distance, similarity)
			fmt.Printf("[SUGGESTIONS] Found %d similar tags:\n", len(semanticErr.Suggestions))
			for i, tag := range semanticErr.Suggestions {
				fmt.Printf("  [%d] %s\n", i+1, tag.Name)
			}

			// Build suggestions list
			suggestions := make([]dto.TagResponse, 0, len(semanticErr.Suggestions))
			for _, tag := range semanticErr.Suggestions {
				suggestions = append(suggestions, dto.TagResponse{
					ID:   tag.ID.String(),
					Name: tag.Name,
				})
			}

			// Return 409 Conflict with existing tag data and suggestions
			c.JSON(http.StatusConflict, dto.TagDuplicateResponse{
				ExistingTag: dto.TagResponse{
					ID:   semanticErr.ExistingTag.ID.String(),
					Name: semanticErr.ExistingTag.Name,
				},
				Message:     "Did you mean '" + semanticErr.ExistingTag.Name + "'?",
				Similarity:  similarity,
				Suggestions: suggestions,
			})
			fmt.Printf("[RESPONSE] 409 Conflict (%dms)\n", time.Since(startTime).Milliseconds())
			return
		}

		// Handle other errors
		statusCode := http.StatusInternalServerError
		if err.Error() != "" && len(err.Error()) >= 3 && err.Error()[:3] == "tag" {
			statusCode = http.StatusConflict
		}
		fmt.Printf("[ERROR] CreateTag failed: %v (status: %d)\n", err, statusCode)
		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to create tag",
			Message: err.Error(),
			Code:    statusCode,
		})
		fmt.Printf("[RESPONSE] %d Error (%dms)\n", statusCode, time.Since(startTime).Milliseconds())
		return
	}

	fmt.Printf("[SUCCESS] Tag created: '%s' (ID: %s) (%dms)\n", tag.Name, tag.ID, time.Since(startTime).Milliseconds())
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

	response, err := h.service.ListTags(c.Request.Context(), req)
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

	tag, err := h.service.AddTagToVideo(c.Request.Context(), videoID, req)
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
*/

// ============================================================
// Canonical-Alias Architecture Handlers (New)
// ============================================================

// AddCanonicalTagToVideo godoc
// @Summary Add canonical tag to video (with auto-resolution)
// @Description Add a tag to a video using 4-layer resolution. Automatically merges similar tags.
// @Description No 409 conflict errors - system handles duplicates transparently.
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Param request body dto.AddVideoTagRequest true "Tag info (provide tag_id OR tag_name)"
// @Success 200 {object} dto.CanonicalTagResponse "Tag added successfully (may be auto-merged)"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /v2/mod/videos/{id}/tags [post]
func (h *TagHandler) AddCanonicalTagToVideo(c *gin.Context) {
	startTime := time.Now()
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

	fmt.Printf("[ADD_CANONICAL_TAG] Video: %s, Input: %v\n", videoID, req)

	tag, err := h.serviceV2.AddCanonicalTagToVideo(c.Request.Context(), videoID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()

		var apiResponse dto.APIResponse
		if errMsg == "video not found" {
			statusCode = http.StatusNotFound
			apiResponse = dto.NewNotFoundResponse("video", videoID)
		} else if errMsg == "canonical tag not found" {
			statusCode = http.StatusNotFound
			tagID := ""
			if req.TagID != nil {
				tagID = *req.TagID
			}
			apiResponse = dto.NewNotFoundResponse("canonical tag", tagID)
		} else if len(errMsg) >= 6 && (errMsg[:6] == "either" || errMsg[:7] == "invalid") {
			statusCode = http.StatusBadRequest
			apiResponse = dto.NewValidationErrorResponse("request", errMsg)
		} else {
			apiResponse = dto.NewInternalErrorResponse(errMsg)
		}

		fmt.Printf("[ERROR] AddCanonicalTagToVideo failed: %v (status: %d, %dms)\n",
			err, statusCode, time.Since(startTime).Milliseconds())
		c.JSON(statusCode, apiResponse)
		return
	}

	// Build tag response data
	tagData := dto.TagResolveResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}

	// If tag_name was provided, add matched_alias for UI feedback
	if req.TagName != nil {
		tagData.MatchedAlias = req.TagName
	}

	// Build metadata
	processingTime := time.Since(startTime).Milliseconds()
	metadata := dto.NewTagMetadata(dto.TagMetadataOptions{
		ProcessingTimeMs: &processingTime,
	})

	apiResponse := dto.NewSuccessResponse(
		tagData,
		fmt.Sprintf("Tag '%s' added to video successfully", tag.Name),
		metadata,
	)

	fmt.Printf("[SUCCESS] Canonical tag '%s' added to video %s (%dms)\n",
		tag.Name, videoID, processingTime)
	c.JSON(http.StatusOK, apiResponse)
}

// RemoveCanonicalTagFromVideo godoc
// @Summary Remove canonical tag from video
// @Description Remove a canonical tag from a video
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Param tag_id path string true "Canonical Tag ID (UUID)"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /v2/mod/videos/{id}/tags/{tag_id} [delete]
func (h *TagHandler) RemoveCanonicalTagFromVideo(c *gin.Context) {
	videoID := c.Param("id")
	tagID := c.Param("tag_id")

	if err := h.serviceV2.RemoveCanonicalTagFromVideo(c.Request.Context(), videoID, tagID); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Failed to remove canonical tag",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetVideoCanonicalTags godoc
// @Summary Get video canonical tags
// @Description Get all canonical tags for a video
// @Tags Videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID)"
// @Success 200 {array} dto.TagResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /v2/mod/videos/{id}/tags [get]
func (h *TagHandler) GetVideoCanonicalTags(c *gin.Context) {
	videoID := c.Param("id")

	tags, err := h.serviceV2.GetVideoCanonicalTags(c.Request.Context(), videoID)
	if err != nil {
		apiResponse := dto.NewNotFoundResponse("video", videoID)
		c.JSON(http.StatusNotFound, apiResponse)
		return
	}

	apiResponse := dto.NewSuccessResponse(tags, fmt.Sprintf("Video has %d tags", len(tags)), nil)
	c.JSON(http.StatusOK, apiResponse)
}

// ListCanonicalTags godoc
// @Summary List canonical tags
// @Description Get paginated list of canonical tags with optional search
// @Tags Tags
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param query query string false "Search query"
// @Success 200 {object} dto.TagListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /v2/mod/tags [get]
func (h *TagHandler) ListCanonicalTags(c *gin.Context) {
	var req dto.TagListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apiResponse := dto.NewValidationErrorResponse("query", err.Error())
		c.JSON(http.StatusBadRequest, apiResponse)
		return
	}

	response, err := h.serviceV2.ListCanonicalTags(c.Request.Context(), req)
	if err != nil {
		apiResponse := dto.NewInternalErrorResponse("Failed to list canonical tags: " + err.Error())
		c.JSON(http.StatusInternalServerError, apiResponse)
		return
	}

	// Wrap list response in standardized format
	metadata := &dto.Metadata{
		Pagination: &response.Pagination,
	}
	apiResponse := dto.NewSuccessResponse(response.Data, "Tags retrieved successfully", metadata)
	c.JSON(http.StatusOK, apiResponse)
}

// GetCanonicalTag godoc
// @Summary Get canonical tag by ID
// @Description Get canonical tag details by tag ID
// @Tags Tags
// @Accept json
// @Produce json
// @Param id path string true "Canonical Tag ID (UUID)"
// @Success 200 {object} dto.TagResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /v2/mod/tags/{id} [get]
func (h *TagHandler) GetCanonicalTag(c *gin.Context) {
	id := c.Param("id")

	tag, err := h.serviceV2.GetCanonicalTagByID(c.Request.Context(), id)
	if err != nil {
		apiResponse := dto.NewNotFoundResponse("canonical tag", id)
		c.JSON(http.StatusNotFound, apiResponse)
		return
	}

	apiResponse := dto.NewSuccessResponse(tag, "Tag retrieved successfully", nil)
	c.JSON(http.StatusOK, apiResponse)
}

// CreateCanonicalTag godoc
// @Summary Create a new canonical tag (v2 - with auto-resolution)
// @Description Create a canonical tag using 4-layer resolution. Automatically merges similar tags.
// @Description Returns existing tag if semantically similar (>85% similarity).
// @Tags Tags
// @Accept json
// @Produce json
// @Param tag body dto.CreateTagRequest true "Tag data"
// @Success 201 {object} dto.CanonicalTagResponse "Tag created"
// @Success 200 {object} dto.CanonicalTagResponse "Existing similar tag returned (auto-merged)"
// @Failure 400 {object} dto.ErrorResponse
// @Router /v2/mod/tags [post]
func (h *TagHandler) CreateCanonicalTag(c *gin.Context) {
	startTime := time.Now()
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[ERROR] Invalid JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	fmt.Printf("[CREATE_CANONICAL_TAG] Received: '%s'\n", req.Name)

	// Use ResolveTag to handle 4-layer resolution
	canonicalTag, matchedAlias, isNewTag, err := h.serviceV2.ResolveTag(c.Request.Context(), req.Name)
	if err != nil {
		statusCode := http.StatusInternalServerError
		fmt.Printf("[ERROR] CreateCanonicalTag failed: %v (status: %d)\n", err, statusCode)
		apiResponse := dto.NewInternalErrorResponse("Failed to create canonical tag: " + err.Error())
		c.JSON(statusCode, apiResponse)
		return
	}

	// Build tag response data
	tagData := dto.TagResolveResponse{
		ID:           canonicalTag.ID.String(),
		Name:         canonicalTag.DisplayName,
		MatchedAlias: &matchedAlias,
	}

	// Build metadata
	processingTime := time.Since(startTime).Milliseconds()
	originalInput := req.Name
	canonicalName := canonicalTag.DisplayName

	metadata := dto.NewTagMetadata(dto.TagMetadataOptions{
		IsNewResource:    &isNewTag,
		OriginalInput:    &originalInput,
		CanonicalName:    &canonicalName,
		ProcessingTimeMs: &processingTime,
	})

	// Return appropriate response based on operation result
	var apiResponse dto.APIResponse
	var statusCode int

	if isNewTag {
		// New canonical tag created
		statusCode = http.StatusCreated
		apiResponse = dto.NewCreatedResponse(
			tagData,
			fmt.Sprintf("Tag '%s' created successfully", canonicalTag.DisplayName),
			metadata,
		)
		fmt.Printf("[SUCCESS] New canonical tag created: '%s' (ID: %s) (%dms)\n",
			canonicalTag.DisplayName, canonicalTag.ID, processingTime)
	} else {
		// Auto-merged to existing tag
		statusCode = http.StatusOK
		autoMerged := true
		mergedIntoID := canonicalTag.ID.String()
		metadata.AutoMerged = &autoMerged
		metadata.MergedInto = &mergedIntoID

		apiResponse = dto.NewMergedResponse(
			tagData,
			fmt.Sprintf("Tag auto-merged with existing '%s'", canonicalTag.DisplayName),
			metadata,
		)
		fmt.Printf("[SUCCESS] Auto-merged to existing tag: '%s' via alias '%s' (ID: %s) (%dms)\n",
			canonicalTag.DisplayName, matchedAlias, canonicalTag.ID, processingTime)
	}

	c.JSON(statusCode, apiResponse)
}

// SearchCanonicalTags godoc
// @Summary Search canonical tags
// @Description Search canonical tags by name (hybrid: SQL LIKE + Vector similarity)
// @Tags Tags
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Max results" default(10)
// @Success 200 {array} dto.TagResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /v2/mod/tags/search [get]
func (h *TagHandler) SearchCanonicalTags(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		apiResponse := dto.NewValidationErrorResponse("q", "Query parameter 'q' is required")
		c.JSON(http.StatusBadRequest, apiResponse)
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	tags, err := h.serviceV2.SearchCanonicalTags(c.Request.Context(), query, limit)
	if err != nil {
		apiResponse := dto.NewInternalErrorResponse("Failed to search canonical tags: " + err.Error())
		c.JSON(http.StatusInternalServerError, apiResponse)
		return
	}

	apiResponse := dto.NewSuccessResponse(tags, fmt.Sprintf("Found %d tags", len(tags)), nil)
	c.JSON(http.StatusOK, apiResponse)
}

// ============================================================
// Tag Merge Operations
// ============================================================

// MergeTags godoc
// @Summary Manually merge tags
// @Description Manually merge source tag into target tag. Source becomes an alias of target.
// @Description All aliases and video relationships are transferred to target.
// @Description Source canonical tag is deleted after merge.
// @Tags Tags
// @Accept json
// @Produce json
// @Param request body dto.MergeTagsRequest true "Source and target tag IDs"
// @Success 200 {object} dto.APIResponse "Merge successful"
// @Failure 400 {object} dto.APIResponse "Invalid request or tags are the same"
// @Failure 404 {object} dto.APIResponse "Source or target tag not found"
// @Failure 500 {object} dto.APIResponse "Internal error"
// @Router /v2/mod/tags/merge [post]
func (h *TagHandler) MergeTags(c *gin.Context) {
	startTime := time.Now()

	var req dto.MergeTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiResponse := dto.NewValidationErrorResponse("request", err.Error())
		c.JSON(http.StatusBadRequest, apiResponse)
		return
	}

	fmt.Printf("[MERGE_TAGS] Request: source=%s, target=%s\n", req.SourceID, req.TargetID)

	// Call service to perform merge
	result, err := h.serviceV2.MergeTags(c.Request.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()

		var apiResponse dto.APIResponse

		// Handle specific error cases
		if errMsg == "source tag not found" {
			statusCode = http.StatusNotFound
			apiResponse = dto.NewNotFoundResponse("source tag", req.SourceID)
		} else if errMsg == "target tag not found" {
			statusCode = http.StatusNotFound
			apiResponse = dto.NewNotFoundResponse("target tag", req.TargetID)
		} else if errMsg == "source and target tags must be different" {
			statusCode = http.StatusBadRequest
			apiResponse = dto.NewValidationErrorResponse("source_id", "Source and target tags must be different")
		} else {
			apiResponse = dto.NewInternalErrorResponse("Failed to merge tags: " + errMsg)
		}

		fmt.Printf("[MERGE_TAGS] ✗ Failed: %v (status: %d, %dms)\n",
			err, statusCode, time.Since(startTime).Milliseconds())
		c.JSON(statusCode, apiResponse)
		return
	}

	// Build metadata
	processingTime := time.Since(startTime).Milliseconds()
	metadata := dto.NewTagMetadata(dto.TagMetadataOptions{
		ProcessingTimeMs: &processingTime,
	})

	// Build success response
	apiResponse := dto.NewSuccessResponse(
		result,
		fmt.Sprintf("Successfully merged %d aliases into '%s'", result.MergedAliasCount, result.TargetTag.Name),
		metadata,
	)

	fmt.Printf("[MERGE_TAGS] ✓ Success: Merged %d aliases (%dms)\n",
		result.MergedAliasCount, processingTime)
	c.JSON(http.StatusOK, apiResponse)
}

// ============================================================
// Tag Approval Operations
// ============================================================

// UpdateTagApproval godoc
// @Summary Update tag approval status
// @Description Update the approval status of a canonical tag
// @Tags Tags
// @Accept json
// @Produce json
// @Param id path string true "Canonical Tag ID (UUID)"
// @Param request body dto.UpdateTagApprovalRequest true "Approval status"
// @Success 200 {object} dto.APIResponse "Update successful"
// @Failure 400 {object} dto.APIResponse "Invalid request"
// @Failure 404 {object} dto.APIResponse "Tag not found"
// @Failure 500 {object} dto.APIResponse "Internal error"
// @Router /v2/mod/tags/{id}/approve [patch]
func (h *TagHandler) UpdateTagApproval(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTagApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiResponse := dto.NewValidationErrorResponse("request", err.Error())
		c.JSON(http.StatusBadRequest, apiResponse)
		return
	}

	fmt.Printf("[UPDATE_TAG_APPROVAL] Tag ID: %s, IsApproved: %v\n", id, req.IsApproved)

	tag, err := h.serviceV2.UpdateTagApproval(c.Request.Context(), id, req.IsApproved)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "tag not found" || errMsg == "invalid tag ID" {
			apiResponse := dto.NewNotFoundResponse("tag", id)
			c.JSON(http.StatusNotFound, apiResponse)
			return
		}
		apiResponse := dto.NewInternalErrorResponse("Failed to update tag approval: " + errMsg)
		c.JSON(http.StatusInternalServerError, apiResponse)
		return
	}

	apiResponse := dto.NewSuccessResponse(
		tag,
		fmt.Sprintf("Tag '%s' approval status updated to %v", tag.Name, req.IsApproved),
		nil,
	)

	fmt.Printf("[UPDATE_TAG_APPROVAL] ✓ Success: Tag '%s' is now %s\n",
		tag.Name, map[bool]string{true: "approved", false: "unapproved"}[req.IsApproved])
	c.JSON(http.StatusOK, apiResponse)
}
