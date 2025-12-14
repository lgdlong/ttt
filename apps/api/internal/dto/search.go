package dto

// TranscriptSearchRequest - Deep search in transcripts
type TranscriptSearchRequest struct {
	Query string `form:"q" binding:"required,min=2"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=50" default:"20"`
}

// TranscriptSearchResult - Result for transcript deep search
type TranscriptSearchResult struct {
	VideoID      string  `json:"video_id"`
	VideoTitle   string  `json:"video_title"`
	ThumbnailURL string  `json:"thumbnail_url"`
	StartTime    int     `json:"start_time"`
	EndTime      int     `json:"end_time"`
	Text         string  `json:"text"`
	Rank         float64 `json:"rank"` // Relevance score
}

// TranscriptSearchResponse - Response with search results
type TranscriptSearchResponse struct {
	Query   string                   `json:"query"`
	Results []TranscriptSearchResult `json:"results"`
	Total   int                      `json:"total"`
}

// TagSearchRequest - Semantic search for tags using vector similarity
type TagSearchRequest struct {
	Query string `form:"q" binding:"required,min=2"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=10" default:"5"`
}

// TagSearchResult - Tag with similarity score
type TagSearchResult struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Similarity float64 `json:"similarity"` // Cosine similarity (0-1)
}

// TagSearchResponse - Response with tag search results
type TagSearchResponse struct {
	Query   string            `json:"query"`
	Results []TagSearchResult `json:"results"`
	Total   int               `json:"total"`
}
