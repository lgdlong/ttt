package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/repository"
	"fmt"
	"math"

	"github.com/google/uuid"
)

type VideoService interface {
	GetVideoList(req dto.ListVideoRequest) (*dto.VideoListResponse, error)
	GetVideoDetail(id string) (*dto.VideoDetailResponse, error)
	GetVideoTranscript(id string) (*dto.TranscriptResponse, error)
	SearchTranscripts(req dto.TranscriptSearchRequest) (*dto.TranscriptSearchResponse, error)
	SearchTags(req dto.TagSearchRequest) (*dto.TagSearchResponse, error)
}

type videoService struct {
	repo repository.VideoRepository
}

func NewVideoService(repo repository.VideoRepository) VideoService {
	return &videoService{repo: repo}
}

// GetVideoList retrieves paginated video list
func (s *videoService) GetVideoList(req dto.ListVideoRequest) (*dto.VideoListResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	videos, total, err := s.repo.GetVideoList(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get video list: %w", err)
	}

	// Convert to DTOs
	videoCards := make([]dto.VideoCardResponse, len(videos))
	for i, video := range videos {
		videoCards[i] = s.toVideoCardResponse(&video)
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &dto.VideoListResponse{
		Data: videoCards,
		Pagination: dto.PaginationMetadata{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetVideoDetail retrieves single video with full details
func (s *videoService) GetVideoDetail(id string) (*dto.VideoDetailResponse, error) {
	videoUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid video id: %w", err)
	}

	video, err := s.repo.GetVideoByID(videoUUID)
	if err != nil {
		return nil, fmt.Errorf("video not found: %w", err)
	}

	return s.toVideoDetailResponse(video), nil
}

// GetVideoTranscript retrieves transcript segments for a video
func (s *videoService) GetVideoTranscript(id string) (*dto.TranscriptResponse, error) {
	videoUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid video id: %w", err)
	}

	segments, err := s.repo.GetVideoTranscript(videoUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript: %w", err)
	}

	return s.toTranscriptResponse(id, segments), nil
}

// SearchTranscripts performs full-text search on transcripts
func (s *videoService) SearchTranscripts(req dto.TranscriptSearchRequest) (*dto.TranscriptSearchResponse, error) {
	if req.Limit < 1 {
		req.Limit = 20
	}

	results, err := s.repo.SearchTranscripts(req.Query, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("transcript search failed: %w", err)
	}

	return &dto.TranscriptSearchResponse{
		Query:   req.Query,
		Results: results,
		Total:   len(results),
	}, nil
}

// SearchTags performs semantic search on tags using vector similarity
func (s *videoService) SearchTags(req dto.TagSearchRequest) (*dto.TagSearchResponse, error) {
	if req.Limit < 1 {
		req.Limit = 5
	}

	// TODO: Implement embedding generation for query
	// For now, return empty results with TODO message
	// In production, you would:
	// 1. Call OpenAI/Azure OpenAI embedding API with req.Query
	// 2. Get back embedding vector ([]float32)
	// 3. Pass to repo.SearchTagsByVector

	return &dto.TagSearchResponse{
		Query:   req.Query,
		Results: []dto.TagSearchResult{},
		Total:   0,
	}, fmt.Errorf("TODO: implement embedding generation service")

	// Example implementation (when embedding service is ready):
	/*
		embedding, err := s.embeddingService.GenerateEmbedding(req.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding: %w", err)
		}

		results, err := s.repo.SearchTagsByVector(embedding, req.Limit, 0.7)
		if err != nil {
			return nil, fmt.Errorf("tag search failed: %w", err)
		}

		return &dto.TagSearchResponse{
			Query:   req.Query,
			Results: results,
			Total:   len(results),
		}, nil
	*/
}

// Helper: Convert domain.Video to dto.VideoCardResponse
func (s *videoService) toVideoCardResponse(video *domain.Video) dto.VideoCardResponse {
	return dto.VideoCardResponse{
		ID:            video.ID.String(),
		YoutubeID:     video.YoutubeID,
		Title:         video.Title,
		ThumbnailURL:  video.ThumbnailURL,
		Duration:      video.Duration,
		PublishedAt:   video.PublishedAt.Format("2006-01-02"),
		ViewCount:     video.ViewCount,
		HasTranscript: len(video.Segments) > 0,
	}
}

// Helper: Convert domain.Video to dto.VideoDetailResponse
func (s *videoService) toVideoDetailResponse(video *domain.Video) *dto.VideoDetailResponse {
	tags := make([]dto.TagResponse, len(video.Tags))
	for i, tag := range video.Tags {
		tags[i] = dto.TagResponse{
			ID:   tag.ID.String(),
			Name: tag.Name,
		}
	}

	return &dto.VideoDetailResponse{
		VideoCardResponse: s.toVideoCardResponse(video),
		Tags:              tags,
	}
}

// Helper: Convert segments to dto.TranscriptResponse
func (s *videoService) toTranscriptResponse(videoID string, segments []domain.TranscriptSegment) *dto.TranscriptResponse {
	segmentDTOs := make([]dto.SegmentResponse, len(segments))
	for i, seg := range segments {
		segmentDTOs[i] = dto.SegmentResponse{
			ID:        seg.ID,
			StartTime: seg.StartTime,
			EndTime:   seg.EndTime,
			Text:      seg.TextContent,
		}
	}

	return &dto.TranscriptResponse{
		VideoID:  videoID,
		Segments: segmentDTOs,
	}
}
