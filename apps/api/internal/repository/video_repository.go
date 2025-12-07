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
	GetVideoByID(id uuid.UUID) (*domain.Video, error)
	GetVideoTranscript(videoID uuid.UUID) ([]domain.TranscriptSegment, error)

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
