package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/repository"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

type videoService struct {
	repo repository.VideoRepository
}

func NewVideoService(repo repository.VideoRepository) VideoService {
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
		videoCards[i] = s.toVideoCardResponse(&video)
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
		HasTranscript: video.HasTranscript,
	}
}

// Helper: Convert domain.Video to dto.VideoDetailResponse
func (s *videoService) toVideoDetailResponse(video *domain.Video) *dto.VideoDetailResponse {
	tags := make([]dto.TagResponse, len(video.CanonicalTags))
	for i, tag := range video.CanonicalTags {
		tags[i] = dto.TagResponse{
			ID:   tag.ID.String(),
			Name: tag.DisplayName,
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

// CreateVideo creates a new video by fetching metadata from YouTube
func (s *videoService) CreateVideo(req dto.CreateVideoRequest) (*dto.VideoCreateResponse, error) {
	// Check if video already exists
	if existing, _ := s.repo.GetVideoByYoutubeID(req.YoutubeID); existing != nil {
		return nil, fmt.Errorf("video with YouTube ID '%s' already exists", req.YoutubeID)
	}

	// Fetch metadata from YouTube
	youtubeInfo, err := s.fetchYouTubeMetadata(req.YoutubeID)
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
	metadata, err := s.fetchYouTubeMetadata(youtubeID)
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
		videoCards[i] = s.toVideoCardResponse(&video)
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

// fetchYouTubeMetadata fetches video metadata from YouTube Data API
func (s *videoService) fetchYouTubeMetadata(youtubeID string) (*dto.YouTubeVideoInfo, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		// Fallback: create placeholder metadata if no API key
		return &dto.YouTubeVideoInfo{
			ID:           youtubeID,
			Title:        fmt.Sprintf("Video %s (pending metadata)", youtubeID),
			PublishedAt:  time.Now(),
			Duration:     0,
			ViewCount:    0,
			ThumbnailURL: fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", youtubeID),
		}, nil
	}

	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails,statistics&id=%s&key=%s",
		youtubeID, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call YouTube API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				PublishedAt string `json:"publishedAt"`
				Thumbnails  struct {
					High struct {
						URL string `json:"url"`
					} `json:"high"`
					Maxres struct {
						URL string `json:"url"`
					} `json:"maxres"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"` // ISO 8601 format: PT1H2M3S
			} `json:"contentDetails"`
			Statistics struct {
				ViewCount string `json:"viewCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode YouTube response: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("video not found on YouTube")
	}

	item := result.Items[0]

	// Parse published date
	publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
	if err != nil {
		publishedAt = time.Now()
	}

	// Parse duration (ISO 8601 to seconds)
	duration := parseDuration(item.ContentDetails.Duration)

	// Parse view count
	viewCount, _ := strconv.Atoi(item.Statistics.ViewCount)

	// Get best thumbnail
	thumbnailURL := item.Snippet.Thumbnails.High.URL
	if item.Snippet.Thumbnails.Maxres.URL != "" {
		thumbnailURL = item.Snippet.Thumbnails.Maxres.URL
	}

	return &dto.YouTubeVideoInfo{
		ID:           item.ID,
		Title:        item.Snippet.Title,
		PublishedAt:  publishedAt,
		Duration:     duration,
		ViewCount:    viewCount,
		ThumbnailURL: thumbnailURL,
	}, nil
}

// parseDuration converts ISO 8601 duration (PT1H2M3S) to seconds
func parseDuration(isoDuration string) int {
	// Remove PT prefix
	d := strings.TrimPrefix(isoDuration, "PT")

	var hours, minutes, seconds int

	// Match hours
	if h := regexp.MustCompile(`(\d+)H`).FindStringSubmatch(d); len(h) > 1 {
		hours, _ = strconv.Atoi(h[1])
	}
	// Match minutes
	if m := regexp.MustCompile(`(\d+)M`).FindStringSubmatch(d); len(m) > 1 {
		minutes, _ = strconv.Atoi(m[1])
	}
	// Match seconds
	if s := regexp.MustCompile(`(\d+)S`).FindStringSubmatch(d); len(s) > 1 {
		seconds, _ = strconv.Atoi(s[1])
	}

	return hours*3600 + minutes*60 + seconds
}
