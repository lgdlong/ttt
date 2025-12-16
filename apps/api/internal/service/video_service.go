package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/helper"
	"fmt"
	"math"

	"github.com/google/uuid"
)

type videoService struct {
	repo domain.VideoRepository
}

func NewVideoService(repo domain.VideoRepository) domain.VideoService {
	return &videoService{repo: repo}
}

// GetModVideoList retrieves videos for mod dashboard with tags and optional has_transcript filter
func (s *videoService) GetModVideoList(page, pageSize int, searchQuery, tagIDsStr, hasTranscriptStr string) ([]dto.ModVideoResponse, int64, error) {
	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	// Get videos from repository
	videos, total, err := s.repo.GetModVideoList(offset, pageSize, searchQuery, tagIDsStr, hasTranscriptStr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get video list: %w", err)
	}

	// Get review counts for all videos in this page
	videoIDs := make([]uuid.UUID, len(videos))
	for i, video := range videos {
		videoIDs[i] = video.ID
	}

	reviewCounts, err := s.repo.GetReviewCountsForVideos(videoIDs)
	if err != nil {
		// Log error but don't fail - just set review counts to 0
		fmt.Printf("Warning: Failed to get review counts: %v\n", err)
		reviewCounts = make(map[uuid.UUID]int)
	}

	// Convert to mod DTOs with tags
	result := make([]dto.ModVideoResponse, len(videos))
	for i, video := range videos {
		tags := make([]dto.TagResponse, len(video.CanonicalTags))
		for j, tag := range video.CanonicalTags {
			tags[j] = dto.TagResponse{
				ID:   tag.ID.String(),
				Name: tag.DisplayName,
			}
		}

		result[i] = dto.ModVideoResponse{
			ID:            video.ID.String(),
			YoutubeID:     video.YoutubeID,
			Title:         video.Title,
			Description:   "", // TODO: Add description field to Video model if needed
			ThumbnailURL:  video.ThumbnailURL,
			Duration:      video.Duration,
			PublishedAt:   video.PublishedAt.Format("2006-01-02"),
			ViewCount:     video.ViewCount,
			HasTranscript: video.HasTranscript,
			ReviewCount:   reviewCounts[video.ID],
			Tags:          tags,
			CreatedAt:     video.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     video.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return result, total, nil
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

	// Get review counts for all videos in this page
	videoIDs := make([]uuid.UUID, len(videos))
	for i, video := range videos {
		videoIDs[i] = video.ID
	}

	reviewCounts, err := s.repo.GetReviewCountsForVideos(videoIDs)
	if err != nil {
		// Log error but don't fail - just set review counts to 0
		fmt.Printf("Warning: Failed to get review counts: %v\n", err)
		reviewCounts = make(map[uuid.UUID]int)
	}

	// Convert to DTOs with review counts
	videoCards := make([]dto.VideoCardResponse, len(videos))
	for i, video := range videos {
		videoCards[i] = helper.ToVideoCardResponse(&video)
		videoCards[i].ReviewCount = reviewCounts[video.ID]
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

	// Fetch review count for this single video
	reviewCounts, err := s.repo.GetReviewCountsForVideos([]uuid.UUID{videoUUID})
	reviewCount := 0
	if err == nil {
		reviewCount = reviewCounts[videoUUID]
	}

	return helper.ToVideoDetailResponse(video, reviewCount), nil
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

	return helper.ToTranscriptResponse(id, segments), nil
}

// UpdateSegment updates a single transcript segment
func (s *videoService) UpdateSegment(id uint, req dto.UpdateSegmentRequest) (*dto.SegmentResponse, error) {
	segment, err := s.repo.UpdateSegment(id, req.TextContent)
	if err != nil {
		return nil, fmt.Errorf("failed to update segment: %w", err)
	}

	return &dto.SegmentResponse{
		ID:        segment.ID,
		StartTime: segment.StartTime,
		EndTime:   segment.EndTime,
		Text:      segment.TextContent,
	}, nil
}

// CreateSegment creates a new transcript segment
func (s *videoService) CreateSegment(videoID string, req dto.CreateSegmentRequest) (*dto.SegmentResponse, error) {
	videoUUID, err := uuid.Parse(videoID)
	if err != nil {
		return nil, fmt.Errorf("invalid video id: %w", err)
	}

	segment, err := s.repo.CreateSegment(videoUUID, req.StartTime, req.EndTime, req.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to create segment: %w", err)
	}

	return &dto.SegmentResponse{
		ID:        segment.ID,
		StartTime: segment.StartTime,
		EndTime:   segment.EndTime,
		Text:      segment.TextContent,
	}, nil
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

// CreateVideo creates a new video by fetching metadata from YouTube
func (s *videoService) CreateVideo(req dto.CreateVideoRequest) (*dto.VideoCreateResponse, error) {
	// Check if video already exists
	if existing, _ := s.repo.GetVideoByYoutubeID(req.YoutubeID); existing != nil {
		return nil, fmt.Errorf("video with YouTube ID '%s' already exists", req.YoutubeID)
	}

	// Fetch metadata from YouTube
	youtubeInfo, err := helper.FetchYouTubeMetadata(req.YoutubeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch YouTube metadata: %w", err)
	}

	// Create video record
	video := &domain.Video{
		YoutubeID:     req.YoutubeID,
		Title:         youtubeInfo.Title,
		PublishedAt:   youtubeInfo.PublishedAt,
		Duration:      youtubeInfo.Duration,
		ViewCount:     youtubeInfo.ViewCount,
		ThumbnailURL:  youtubeInfo.ThumbnailURL,
		HasTranscript: false,
	}

	if err := s.repo.Create(video); err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	return &dto.VideoCreateResponse{
		ID:            video.ID.String(),
		YoutubeID:     video.YoutubeID,
		Title:         video.Title,
		PublishedAt:   video.PublishedAt,
		Duration:      video.Duration,
		ViewCount:     video.ViewCount,
		ThumbnailURL:  video.ThumbnailURL,
		HasTranscript: video.HasTranscript,
		CreatedAt:     video.CreatedAt,
	}, nil
}

// DeleteVideo soft deletes a video
// PreviewYouTubeVideo fetches YouTube metadata without saving to database
func (s *videoService) PreviewYouTubeVideo(youtubeID string) (*dto.VideoCreateResponse, error) {
	// Validate YouTube ID format
	if youtubeID == "" || len(youtubeID) != 11 {
		return nil, fmt.Errorf("invalid YouTube ID format")
	}

	// Fetch metadata from YouTube API
	metadata, err := helper.FetchYouTubeMetadata(youtubeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch YouTube metadata: %w", err)
	}

	// Return preview response (no database save)
	return &dto.VideoCreateResponse{
		ID:           "", // Empty since not saved yet
		YoutubeID:    metadata.ID,
		Title:        metadata.Title,
		ThumbnailURL: metadata.ThumbnailURL,
		Duration:     metadata.Duration,
		ViewCount:    metadata.ViewCount,
		PublishedAt:  metadata.PublishedAt,
	}, nil
}

func (s *videoService) DeleteVideo(id string) error {
	videoUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid video ID: %w", err)
	}

	// Check if video exists
	if _, err := s.repo.GetVideoByID(videoUUID); err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	if err := s.repo.Delete(videoUUID); err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	return nil
}

// SearchVideos searches videos by title
func (s *videoService) SearchVideos(query string, page, limit int) (*dto.VideoListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	videos, total, err := s.repo.SearchVideos(query, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search videos: %w", err)
	}

	videoCards := make([]dto.VideoCardResponse, len(videos))
	for i, video := range videos {
		videoCards[i] = helper.ToVideoCardResponse(&video)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.VideoListResponse{
		Data: videoCards,
		Pagination: dto.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}
