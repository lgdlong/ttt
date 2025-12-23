package repository

import (
	"api/internal/domain"
	"api/internal/dto"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type videoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) domain.VideoRepository {
	return &videoRepository{db: db}
}

// GetVideoList retrieves paginated list of videos
func (r *videoRepository) GetVideoList(req dto.ListVideoRequest) ([]domain.Video, int64, error) {
	var videos []domain.Video
	var total int64

	query := r.db.Model(&domain.Video{}).Preload("CanonicalTags")

	// Apply search query if provided (searches in Title OR Tag Name)
	if req.Q != "" {
		searchQuery := "%" + req.Q + "%"
		// LEFT JOIN to canonical_tags for tag name search
		query = query.Joins("LEFT JOIN video_canonical_tags vct ON vct.video_id = videos.id").
			Joins("LEFT JOIN canonical_tags ct ON ct.id = vct.canonical_tag_id").
			Where("LOWER(videos.title) LIKE LOWER(?) OR LOWER(ct.display_name) LIKE LOWER(?)", searchQuery, searchQuery).
			Group("videos.id") // Prevent duplicates when video matches multiple tags
	}

	// Apply tag filter if provided
	if req.TagID != "" {
		tagUUID, err := uuid.Parse(req.TagID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid tag_id format: %w", err)
		}
		// If we already have JOINs from search, we need to add another condition
		if req.Q != "" {
			query = query.Where("vct.canonical_tag_id = ?", tagUUID)
		} else {
			query = query.Joins("JOIN video_canonical_tags ON video_canonical_tags.video_id = videos.id").
				Where("video_canonical_tags.canonical_tag_id = ?", tagUUID)
		}
	}

	// Apply has_transcript filter if provided
	if req.HasTranscript != nil {
		query = query.Where("has_transcript = ?", *req.HasTranscript)
	}

	// Apply is_reviewed filter if provided
	if req.IsReviewed != nil {
		if *req.IsReviewed {
			query = query.Where("EXISTS (?)", r.db.Model(&domain.VideoTranscriptReview{}).
				Select("1").
				Where("video_transcript_reviews.video_id = videos.id").
				Limit(1)) // Limit 1 for EXISTS subquery optimization
		} else {
			query = query.Where("NOT EXISTS (?)", r.db.Model(&domain.VideoTranscriptReview{}).
				Select("1").
				Where("video_transcript_reviews.video_id = videos.id").
				Limit(1)) // Limit 1 for NOT EXISTS subquery optimization
		}
	}

	// Count total before pagination (need to count distinct videos)
	countQuery := query.Session(&gorm.Session{})
	if req.Q != "" {
		// When using GROUP BY, we need to count distinct video IDs
		var countResult int64
		if err := r.db.Raw("SELECT COUNT(*) FROM (?) AS subquery", countQuery.Select("videos.id")).Scan(&countResult).Error; err != nil {
			return nil, 0, err
		}
		total = countResult
	} else {
		if err := countQuery.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	// Apply sorting
	switch req.Sort {
	case "popular", "views":
		query = query.Order("videos.view_count DESC")
	case "newest":
		query = query.Order("videos.published_at DESC")
	default:
		query = query.Order("videos.created_at DESC")
	}

	// Apply pagination
	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// GetReviewCountsForVideos retrieves review counts for a list of video IDs
func (r *videoRepository) GetReviewCountsForVideos(videoIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	type reviewCount struct {
		VideoID uuid.UUID
		Count   int
	}

	var counts []reviewCount
	err := r.db.Model(&domain.VideoTranscriptReview{}).
		Select("video_id, COUNT(*) as count").
		Where("video_id IN ?", videoIDs).
		Group("video_id").
		Scan(&counts).Error

	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int)
	for _, c := range counts {
		result[c.VideoID] = c.Count
	}

	return result, nil
}

// GetVideoByID retrieves single video with tags
func (r *videoRepository) GetVideoByID(id uuid.UUID) (*domain.Video, error) {
	var video domain.Video
	if err := r.db.Preload("CanonicalTags").Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("chapter_order ASC")
	}).First(&video, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

// GetVideoByYoutubeID retrieves a video by YouTube ID
func (r *videoRepository) GetVideoByYoutubeID(youtubeID string) (*domain.Video, error) {
	var video domain.Video
	if err := r.db.Preload("CanonicalTags").Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("chapter_order ASC")
	}).Where("youtube_id = ?", youtubeID).First(&video).Error; err != nil {
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
	if err := baseQuery.Preload("CanonicalTags").
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

// CreateSegment creates a new transcript segment
func (r *videoRepository) CreateSegment(videoID uuid.UUID, startTime, endTime int, text string) (*domain.TranscriptSegment, error) {
	segment := &domain.TranscriptSegment{
		VideoID:     videoID,
		StartTime:   startTime,
		EndTime:     endTime,
		TextContent: text,
	}

	if err := r.db.Create(segment).Error; err != nil {
		return nil, err
	}

	return segment, nil
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
		Preload("CanonicalTags"). // Load canonical tags for each video
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
