package dto

// ListVideoRequest - Request params for listing videos
type ListVideoRequest struct {
	Page          int    `form:"page" binding:"omitempty,min=1" default:"1"`
	Limit         int    `form:"limit" binding:"omitempty,min=1,max=50" default:"10"`
	Sort          string `form:"sort" binding:"omitempty,oneof=newest popular views"`
	TagID         string `form:"tag_id" binding:"omitempty,uuid"`
	HasTranscript *bool  `form:"has_transcript" binding:"omitempty"` // nil = all, true = only with transcript, false = only without
	Q             string `form:"q" binding:"omitempty"`              // Search query - searches in Title OR Tag Name
}

// VideoCardResponse - Lightweight video data for grid/list view
type VideoCardResponse struct {
	ID            string `json:"id"`
	YoutubeID     string `json:"youtube_id"`
	Title         string `json:"title"`
	ThumbnailURL  string `json:"thumbnail_url"`
	Duration      int    `json:"duration"` // Seconds - frontend converts to mm:ss
	PublishedAt   string `json:"published_at"`
	ViewCount     int    `json:"view_count"`
	HasTranscript bool   `json:"has_transcript"` // Show "CC" badge
	ReviewCount   int    `json:"review_count"`   // Number of reviews - show "Đã duyệt" badge if > 0
}

// VideoDetailResponse - Full video data with tags
type VideoDetailResponse struct {
	VideoCardResponse
	Tags []TagResponse `json:"tags"`
}

// TranscriptResponse - Full transcript with segments
type TranscriptResponse struct {
	VideoID  string            `json:"video_id"`
	Segments []SegmentResponse `json:"segments"`
}

// SegmentResponse - Individual transcript segment
type SegmentResponse struct {
	ID        uint   `json:"id"`
	StartTime int    `json:"start_time"` // Milliseconds
	EndTime   int    `json:"end_time"`   // Milliseconds
	Text      string `json:"text"`
}

// UpdateSegmentRequest - Request to update a single segment
type UpdateSegmentRequest struct {
	TextContent string `json:"text_content" binding:"required"`
	StartTime   *int   `json:"start_time" binding:"omitempty"`
	EndTime     *int   `json:"end_time" binding:"omitempty"`
}

// CreateSegmentRequest - Request to create a new segment
type CreateSegmentRequest struct {
	StartTime int    `json:"start_time" binding:"required,min=0"`           // Milliseconds
	EndTime   int    `json:"end_time" binding:"required,gtfield=StartTime"` // Milliseconds, must be > StartTime
	Text      string `json:"text" binding:"required,min=1"`
}

// ============ Video Transcript Review DTOs ============

// SubmitReviewRequest - Request to submit a video transcript review
type SubmitReviewRequest struct {
	// Optional: Add any review metadata if needed in future
	Notes string `json:"notes" binding:"omitempty,max=500"`
}

// VideoTranscriptReviewResponse - Response after submitting a review
type VideoTranscriptReviewResponse struct {
	ID            uint   `json:"id"`
	VideoID       string `json:"video_id"`
	UserID        string `json:"user_id"`
	ReviewedAt    string `json:"reviewed_at"`
	TotalReviews  int    `json:"total_reviews"`  // Total reviews for this video
	VideoStatus   string `json:"video_status"`   // Current video status (e.g., "PUBLISHED")
	PointsAwarded int    `json:"points_awarded"` // Points given to reviewer
	Message       string `json:"message"`        // Human-readable status message
}

// PaginationMetadata - Pagination info for list responses
type PaginationMetadata struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// VideoListResponse - Paginated video list with metadata
type VideoListResponse struct {
	Data       []VideoCardResponse `json:"data"`
	Pagination PaginationMetadata  `json:"pagination"`
}

// ModVideoResponse - Video data for mod dashboard
type ModVideoResponse struct {
	ID            string        `json:"id"`
	YoutubeID     string        `json:"youtube_id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ThumbnailURL  string        `json:"thumbnail_url"`
	Duration      int           `json:"duration"`
	PublishedAt   string        `json:"published_at"`
	ViewCount     int           `json:"view_count"`
	HasTranscript bool          `json:"has_transcript"`
	Tags          []TagResponse `json:"tags"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
}

// ModVideoListResponse - Paginated mod video list
type ModVideoListResponse struct {
	Videos   []ModVideoResponse `json:"videos"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}
