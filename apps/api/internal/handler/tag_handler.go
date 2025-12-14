package handler

import (
	"api/internal/domain"
	"api/internal/dto"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	service   domain.TagService
	serviceV2 domain.TagServiceV2
}

func NewTagHandler(service domain.TagService, serviceV2 domain.TagServiceV2) *TagHandler {
	return &TagHandler{
		service:   service,
		serviceV2: serviceV2,
	}
}

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
// @Param approved_only query bool false "Only return approved tags" default(false)
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

	approvedOnly := c.Query("approved_only") == "true"

	tags, err := h.serviceV2.SearchCanonicalTags(c.Request.Context(), query, limit, approvedOnly)
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
		// Use sentinel errors for robust error handling
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidID) {
			apiResponse := dto.NewNotFoundResponse("tag", id)
			c.JSON(http.StatusNotFound, apiResponse)
			return
		}
		apiResponse := dto.NewInternalErrorResponse("Failed to update tag approval: " + err.Error())
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
