package domain

import (
	"api/internal/dto"
)

type VideoService interface {
	GetVideoList(req dto.ListVideoRequest) (*dto.VideoListResponse, error)
	GetVideoDetail(id string) (*dto.VideoDetailResponse, error)
	GetVideoTranscript(id string) (*dto.TranscriptResponse, error)
	UpdateSegment(id uint, req dto.UpdateSegmentRequest) (*dto.SegmentResponse, error)
	CreateSegment(videoID string, req dto.CreateSegmentRequest) (*dto.SegmentResponse, error)
	SearchTranscripts(req dto.TranscriptSearchRequest) (*dto.TranscriptSearchResponse, error)
	SearchTags(req dto.TagSearchRequest) (*dto.TagSearchResponse, error)

	// Video management (Mod)
	GetModVideoList(page, pageSize int, searchQuery, tagIDsStr, hasTranscriptStr string) ([]dto.ModVideoResponse, int64, error)
	CreateVideo(req dto.CreateVideoRequest) (*dto.VideoCreateResponse, error)
	PreviewYouTubeVideo(youtubeID string) (*dto.VideoCreateResponse, error)
	DeleteVideo(id string) error
	SearchVideos(query string, page, limit int) (*dto.VideoListResponse, error)
}
