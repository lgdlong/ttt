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
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TagListRequest - Request params for listing tags
type TagListRequest struct {
	Page  int    `form:"page" binding:"omitempty,min=1" default:"1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100" default:"20"`
	Query string `form:"query" binding:"omitempty"`
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
