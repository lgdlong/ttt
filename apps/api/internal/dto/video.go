package dto

// ListVideoRequest - Request params for listing videos
type ListVideoRequest struct {
	Page  int    `form:"page" binding:"omitempty,min=1" default:"1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=50" default:"10"`
	Sort  string `form:"sort" binding:"omitempty,oneof=newest popular views"`
	TagID string `form:"tag_id" binding:"omitempty,uuid"`
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
