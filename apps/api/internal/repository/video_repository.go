package repository

import (
	"api/internal/domain"
	"api/internal/dto"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VideoRepository interface {
	// Video operations
	GetVideoList(req dto.ListVideoRequest) ([]domain.Video, int64, error)
	GetModVideoList(offset, limit int, searchQuery, tagIDsStr, hasTranscriptStr string) ([]domain.Video, int64, error)
	GetVideoByID(id uuid.UUID) (*domain.Video, error)
	GetVideoByYoutubeID(youtubeID string) (*domain.Video, error)
	GetVideoTranscript(videoID uuid.UUID) ([]domain.TranscriptSegment, error)
	UpdateSegment(id uint, textContent string) (*domain.TranscriptSegment, error)
	Create(video *domain.Video) error
	Update(video *domain.Video) error
	Delete(id uuid.UUID) error // Soft delete
	SearchVideos(query string, page, limit int) ([]domain.Video, int64, error)

	// Search operations
	SearchTranscripts(query string, limit int) ([]dto.TranscriptSearchResult, error)
	SearchTagsByVector(embedding []float32, limit int, minSimilarity float64) ([]dto.TagSearchResult, error)
}

type videoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepository{db: db}
}

// GetVideoList retrieves paginated list of videos
func (r *videoRepository) GetVideoList(req dto.ListVideoRequest) ([]domain.Video, int64, error) {
	var videos []domain.Video
	var total int64

	query := r.db.Model(&domain.Video{}).Preload("Tags")

	// Apply tag filter if provided
	if req.TagID != "" {
		tagUUID, err := uuid.Parse(req.TagID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid tag_id format: %w", err)
		}
		query = query.Joins("JOIN video_tags ON video_tags.video_id = videos.id").
			Where("video_tags.tag_id = ?", tagUUID)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch req.Sort {
	case "popular", "views":
		query = query.Order("view_count DESC")
	case "newest":
		query = query.Order("published_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// GetVideoByID retrieves single video with tags
func (r *videoRepository) GetVideoByID(id uuid.UUID) (*domain.Video, error) {
	var video domain.Video
	if err := r.db.Preload("Tags").First(&video, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

// GetVideoByYoutubeID retrieves a video by YouTube ID
func (r *videoRepository) GetVideoByYoutubeID(youtubeID string) (*domain.Video, error) {
	var video domain.Video
	if err := r.db.Preload("Tags").Where("youtube_id = ?", youtubeID).First(&video).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

// Create creates a new video
func (r *videoRepository) Create(video *domain.Video) error {
	return r.db.Create(video).Error
}

// Update updates a video
func (r *videoRepository) Update(video *domain.Video) error {
	return r.db.Save(video).Error
}

// Delete soft deletes a video
func (r *videoRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Video{}, "id = ?", id).Error
}

// SearchVideos searches videos by title
func (r *videoRepository) SearchVideos(query string, page, limit int) ([]domain.Video, int64, error) {
	var videos []domain.Video
	var total int64

	baseQuery := r.db.Model(&domain.Video{}).
		Where("LOWER(title) LIKE LOWER(?)", "%"+query+"%")

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := baseQuery.Preload("Tags").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// GetVideoTranscript retrieves all transcript segments for a video
func (r *videoRepository) GetVideoTranscript(videoID uuid.UUID) ([]domain.TranscriptSegment, error) {
	var segments []domain.TranscriptSegment
	if err := r.db.Where("video_id = ?", videoID).
		Order("start_time ASC").
		Find(&segments).Error; err != nil {
		return nil, err
	}
	return segments, nil
}

// UpdateSegment updates a single transcript segment
func (r *videoRepository) UpdateSegment(id uint, textContent string) (*domain.TranscriptSegment, error) {
	var segment domain.TranscriptSegment
	if err := r.db.First(&segment, id).Error; err != nil {
		return nil, err
	}

	segment.TextContent = textContent
	if err := r.db.Save(&segment).Error; err != nil {
		return nil, err
	}

	return &segment, nil
}

// SearchTranscripts performs full-text search on transcript segments using tsvector
func (r *videoRepository) SearchTranscripts(query string, limit int) ([]dto.TranscriptSearchResult, error) {
	var results []dto.TranscriptSearchResult

	sql := `
		SELECT 
			ts.video_id,
			v.title as video_title,
			v.thumbnail_url,
			ts.start_time,
			ts.end_time,
			ts.text_content as text,
			ts_rank(ts.tsv, websearch_to_tsquery('english', ?)) as rank
		FROM 
			transcript_segments ts
		JOIN 
			videos v ON ts.video_id = v.id
		WHERE 
			ts.tsv @@ websearch_to_tsquery('english', ?)
		ORDER BY 
			rank DESC, ts.start_time ASC
		LIMIT ?
	`

	if err := r.db.Raw(sql, query, query, limit).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// SearchTagsByVector performs semantic search on tags using vector similarity
func (r *videoRepository) SearchTagsByVector(embedding []float32, limit int, minSimilarity float64) ([]dto.TagSearchResult, error) {
	var results []dto.TagSearchResult

	// Convert embedding to pgvector format
	embeddingStr := fmt.Sprintf("[%v]", embedding)

	sql := `
		SELECT 
			id, 
			name, 
			1 - (embedding <=> ?::vector) as similarity
		FROM 
			tags
		WHERE 
			1 - (embedding <=> ?::vector) > ?
		ORDER BY 
			embedding <=> ?::vector
		LIMIT ?
	`

	if err := r.db.Raw(sql, embeddingStr, embeddingStr, minSimilarity, embeddingStr, limit).
		Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// GetModVideoList retrieves videos for mod dashboard with tags, search, and filtering
func (r *videoRepository) GetModVideoList(offset, limit int, searchQuery, tagIDsStr, hasTranscriptStr string) ([]domain.Video, int64, error) {
	var videos []domain.Video
	var total int64

	query := r.db.
		Preload("Tags"). // Load tags for each video
		Offset(offset).
		Limit(limit).
		Order("created_at DESC")

	// Apply search filter
	if searchQuery != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+searchQuery+"%")
	}

	// Apply has_transcript filter
	if hasTranscriptStr == "true" {
		query = query.Where("has_transcript = ?", true)
	} else if hasTranscriptStr == "false" {
		query = query.Where("has_transcript = ?", false)
	}
	// If "all" or empty, no filter applied

	// Apply tag filter if provided
	if tagIDsStr != "" {
		// This would require parsing comma-separated tag IDs and joining with video_tags table
		// For now, we'll skip tag filtering in the repository level
		// Frontend can filter on received data if needed
	}

	// Count total before pagination (apply same filters as query)
	countQuery := r.db.Model(&domain.Video{})

	if searchQuery != "" {
		countQuery = countQuery.Where("LOWER(title) LIKE ?", "%"+searchQuery+"%")
	}

	if hasTranscriptStr == "true" {
		countQuery = countQuery.Where("has_transcript = ?", true)
	} else if hasTranscriptStr == "false" {
		countQuery = countQuery.Where("has_transcript = ?", false)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count videos: %w", err)
	}

	// Get paginated results
	if err := query.Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get videos: %w", err)
	}

	return videos, total, nil
}
