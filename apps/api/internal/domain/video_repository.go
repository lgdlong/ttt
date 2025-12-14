package domain

import (
	"api/internal/dto"

	"github.com/google/uuid"
)

type VideoRepository interface {
	// Video operations
	GetVideoList(req dto.ListVideoRequest) ([]Video, int64, error)
	GetModVideoList(offset, limit int, searchQuery, tagIDsStr, hasTranscriptStr string) ([]Video, int64, error)
	GetVideoByID(id uuid.UUID) (*Video, error)
	GetVideoByYoutubeID(youtubeID string) (*Video, error)
	GetVideoTranscript(videoID uuid.UUID) ([]TranscriptSegment, error)
	UpdateSegment(id uint, textContent string) (*TranscriptSegment, error)
	CreateSegment(videoID uuid.UUID, startTime, endTime int, text string) (*TranscriptSegment, error)
	Create(video *Video) error
	Update(video *Video) error
	Delete(id uuid.UUID) error // Soft delete
	SearchVideos(query string, page, limit int) ([]Video, int64, error)
	GetReviewCountsForVideos(videoIDs []uuid.UUID) (map[uuid.UUID]int, error)

	// Search operations
	SearchTranscripts(query string, limit int) ([]dto.TranscriptSearchResult, error)
	SearchTagsByVector(embedding []float32, limit int, minSimilarity float64) ([]dto.TagSearchResult, error)
}
