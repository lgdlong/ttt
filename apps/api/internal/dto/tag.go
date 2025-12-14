package dto

import "time"

// ============ Tag DTOs ============

// CreateTagRequest - Request to create a new tag
type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// UpdateTagRequest - Request to update a tag
type UpdateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// TagResponse - Tag data for API responses
type TagResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	IsApproved bool     `json:"is_approved"`
	Aliases    []string `json:"aliases,omitempty"` // List of alias names for display
}

// CanonicalTagResponse - Canonical tag response with alias metadata
type CanonicalTagResponse struct {
	ID           string  `json:"id"`                      // Canonical tag ID
	Name         string  `json:"name"`                    // Display name (canonical)
	MatchedAlias *string `json:"matched_alias,omitempty"` // Original user input (for UI feedback)
}

// TagDuplicateResponse - Response when a similar tag already exists
type TagDuplicateResponse struct {
	ExistingTag TagResponse   `json:"existing_tag"`
	Message     string        `json:"message"`
	Similarity  float64       `json:"similarity"`  // 0.0 to 1.0
	Suggestions []TagResponse `json:"suggestions"` // Gợi ý các tag gần nhất
}

// TagListRequest - Request params for listing tags
type TagListRequest struct {
	Page         int    `form:"page" binding:"omitempty,min=1" default:"1"`
	Limit        int    `form:"limit" binding:"omitempty,min=1,max=100" default:"20"`
	Query        string `form:"query" binding:"omitempty"`
	ApprovedOnly bool   `form:"approved_only" default:"false"`
}

// TagListResponse - Response with list of tags and pagination
type TagListResponse struct {
	Data       []TagResponse      `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// ============ Video Tag Management DTOs ============

// AddVideoTagRequest - Request to add a tag to a video
type AddVideoTagRequest struct {
	TagID   *string `json:"tag_id" binding:"omitempty"`   // Existing tag ID
	TagName *string `json:"tag_name" binding:"omitempty"` // Create new tag if not exists
}

// RemoveVideoTagRequest - Request to remove a tag from a video
type RemoveVideoTagRequest struct {
	TagID string `json:"tag_id" binding:"required"`
}

// ============ Tag Merge DTOs ============

// MergeTagsRequest - Request to manually merge source tag into target tag
// Source tag becomes an alias pointing to target canonical tag
type MergeTagsRequest struct {
	SourceID string `json:"source_id" binding:"required,uuid"` // Tag to be merged (will become alias)
	TargetID string `json:"target_id" binding:"required,uuid"` // Target canonical tag (will remain)
}

// MergeTagsResponse - Response after merging tags
type MergeTagsResponse struct {
	TargetTag        TagResponse `json:"target_tag"`         // The canonical tag that remains
	MergedAliasCount int         `json:"merged_alias_count"` // Number of aliases moved
	SourceTagDeleted bool        `json:"source_tag_deleted"` // Whether source canonical was deleted
}

// ============ Tag Approve DTOs ============

// UpdateTagApprovalRequest - Request to update tag approval status
type UpdateTagApprovalRequest struct {
	IsApproved bool `json:"is_approved"`
}

// ============ Video Create/Delete DTOs ============

// CreateVideoRequest - Request to create a video from YouTube
type CreateVideoRequest struct {
	YoutubeID string `json:"youtube_id" binding:"required,min=11,max=11"`
}

// YouTubeVideoInfo - YouTube video metadata
type YouTubeVideoInfo struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	PublishedAt  time.Time `json:"published_at"`
	Duration     int       `json:"duration"` // seconds
	ViewCount    int       `json:"view_count"`
	ThumbnailURL string    `json:"thumbnail_url"`
}

// VideoCreateResponse - Response after creating a video
type VideoCreateResponse struct {
	ID            string    `json:"id"`
	YoutubeID     string    `json:"youtube_id"`
	Title         string    `json:"title"`
	PublishedAt   time.Time `json:"published_at"`
	Duration      int       `json:"duration"`
	ViewCount     int       `json:"view_count"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	HasTranscript bool      `json:"has_transcript"`
	CreatedAt     time.Time `json:"created_at"`
}
