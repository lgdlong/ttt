package helper

import (
	"api/internal/domain"
	"api/internal/dto"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ToVideoCardResponse converts a domain.Video to a dto.VideoCardResponse.
func ToVideoCardResponse(video *domain.Video) dto.VideoCardResponse {
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

// ToVideoDetailResponse converts a domain.Video to a dto.VideoDetailResponse.
func ToVideoDetailResponse(video *domain.Video, reviewCount int) *dto.VideoDetailResponse {
	tags := make([]dto.TagResponse, len(video.CanonicalTags))
	for i, tag := range video.CanonicalTags {
		tags[i] = dto.TagResponse{
			ID:   tag.ID.String(),
			Name: tag.DisplayName,
		}
	}

	chapters := make([]dto.ChapterResponse, len(video.Chapters))
	for i, ch := range video.Chapters {
		chapters[i] = dto.ChapterResponse{
			ID:        ch.ID,
			Title:     ch.Title,
			Content:   ch.Content,
			StartTime: ch.StartTime,
		}
	}

	cardResponse := ToVideoCardResponse(video)
	cardResponse.ReviewCount = reviewCount

	return &dto.VideoDetailResponse{
		VideoCardResponse: cardResponse,
		Tags:              tags,
		Summary:           video.Summary,
		Chapters:          chapters,
	}
}

// ToTranscriptResponse converts segments to a dto.TranscriptResponse.
func ToTranscriptResponse(videoID string, segments []domain.TranscriptSegment) *dto.TranscriptResponse {
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

// FetchYouTubeMetadata fetches video metadata from YouTube Data API.
func FetchYouTubeMetadata(youtubeID string) (*dto.YouTubeVideoInfo, error) {
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
	duration := ParseDuration(item.ContentDetails.Duration)

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

// ParseDuration converts ISO 8601 duration (PT1H2M3S) to seconds.
func ParseDuration(isoDuration string) int {
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