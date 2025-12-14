package dto

// AdminStatsResponse represents admin dashboard statistics
type AdminStatsResponse struct {
	TotalUsers  int64 `json:"total_users"`
	ActiveUsers int64 `json:"active_users"` // Users not banned
	TotalVideos int64 `json:"total_videos"`
	TotalTags   int64 `json:"total_tags"`
}

// ModStatsResponse represents moderator dashboard statistics
type ModStatsResponse struct {
	TotalVideos          int64 `json:"total_videos"`
	TotalTags            int64 `json:"total_tags"`
	VideosWithTranscript int64 `json:"videos_with_transcript"`
	VideosAddedToday     int64 `json:"videos_added_today"`
}
