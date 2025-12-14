package repository

import (
	"api/internal/domain"
	"api/internal/infrastructure"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type tagRepository struct {
	db           *gorm.DB
	openAIClient *infrastructure.OpenAIClient
}

func NewTagRepository(db *gorm.DB, openAIClient *infrastructure.OpenAIClient) domain.TagRepository {
	return &tagRepository{
		db:           db,
		openAIClient: openAIClient,
	}
}

// ============================================================
// Translation Layer Implementation
// ============================================================

// TranslateText translates text to English using OpenAI GPT-4o-mini
// Returns empty string if OpenAI client is not available (graceful degradation)
func (r *tagRepository) TranslateText(ctx context.Context, text string) (string, error) {
	if r.openAIClient == nil {
		return "", nil // Skip translation if OpenAI client not available
	}
	return r.openAIClient.TranslateToEnglish(ctx, text)
}

// GetEmbeddingForText generates embedding vector for given text using OpenAI
func (r *tagRepository) GetEmbeddingForText(ctx context.Context, text string) ([]float32, error) {
	if r.openAIClient == nil {
		return nil, fmt.Errorf("OpenAI client not available")
	}

	embedding, err := r.openAIClient.GetEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert pgvector.Vector to []float32
	vectorSlice := make([]float32, len(embedding.Slice()))
	copy(vectorSlice, embedding.Slice())

	return vectorSlice, nil
}

// ============================================================
// Canonical-Alias Architecture Implementation
// ============================================================

// GetCanonicalByAlias finds canonical tag by exact normalized text match (Layer 1 - Cache Hit)
// This is the fastest path: ~10ms, no OpenAI cost
func (r *tagRepository) GetCanonicalByAlias(ctx context.Context, normalizedText string) (*domain.CanonicalTag, error) {
	var alias domain.TagAlias

	// Find alias by normalized text
	err := r.db.WithContext(ctx).
		Where("normalized_text = ?", normalizedText).
		First(&alias).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error, proceed to Layer 2
		}
		return nil, fmt.Errorf("failed to query alias: %w", err)
	}

	// Load canonical tag
	var canonical domain.CanonicalTag
	err = r.db.WithContext(ctx).
		Where("id = ?", alias.CanonicalTagID).
		First(&canonical).Error

	if err != nil {
		return nil, fmt.Errorf("failed to load canonical tag: %w", err)
	}

	return &canonical, nil
}

// GetClosestCanonical finds most similar canonical tag using vector search (Layer 3)
// Returns (canonical, score, error)
// If no match above threshold, returns (nil, 0, nil)
func (r *tagRepository) GetClosestCanonical(ctx context.Context, embedding pgvector.Vector, threshold float64) (*domain.CanonicalTag, float64, error) {
	if r.openAIClient == nil {
		return nil, 0, nil // No OpenAI = no semantic search
	}

	type resultRow struct {
		ID             uuid.UUID
		CanonicalTagID uuid.UUID
		RawText        string
		Distance       float64
	}

	var result resultRow

	// Find closest alias using vector similarity
	// cosine distance range: [0, 2]
	// 0 = identical, 1 = orthogonal, 2 = opposite
	// threshold parameter is DISTANCE (not similarity %)
	// Example: threshold=0.30 means distance < 0.30 (85% similarity)
	sqlQuery := `
		SELECT 
			id,
			canonical_tag_id,
			raw_text,
			embedding <=> $1::vector as distance
		FROM tag_aliases
		WHERE embedding IS NOT NULL
			AND embedding <=> $1::vector < $2
		ORDER BY embedding <=> $1::vector ASC
		LIMIT 1
	`

	err := r.db.WithContext(ctx).Raw(sqlQuery, embedding, threshold).Scan(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil // No match found
		}
		return nil, 0, fmt.Errorf("vector search failed: %w", err)
	}

	// If no results returned
	if result.ID == uuid.Nil {
		return nil, 0, nil
	}

	// Convert distance to similarity score (0-1 range)
	// similarity = 1 - (distance / 2)
	// Example: distance=0.3 → similarity=0.85 (85%)
	similarityScore := 1.0 - (result.Distance / 2.0)

	// Load canonical tag
	var canonical domain.CanonicalTag
	err = r.db.WithContext(ctx).
		Where("id = ?", result.CanonicalTagID).
		First(&canonical).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to load canonical tag: %w", err)
	}

	return &canonical, similarityScore, nil
}

// CreateCanonicalTag creates a new canonical tag with its initial alias (atomic transaction)
func (r *tagRepository) CreateCanonicalTag(ctx context.Context, canonical *domain.CanonicalTag, initialAlias *domain.TagAlias) error {
	// Use transaction to ensure atomicity
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Create canonical tag
		if err := tx.Create(canonical).Error; err != nil {
			return fmt.Errorf("failed to create canonical tag: %w", err)
		}

		// Step 2: Set canonical_tag_id in alias
		initialAlias.CanonicalTagID = canonical.ID

		// Step 3: Create initial alias
		if err := tx.Create(initialAlias).Error; err != nil {
			return fmt.Errorf("failed to create initial alias: %w", err)
		}

		return nil
	})
}

// CreateAlias adds a new alias to existing canonical tag
func (r *tagRepository) CreateAlias(ctx context.Context, alias *domain.TagAlias) error {
	return r.db.WithContext(ctx).Create(alias).Error
}

// GetCanonicalByID retrieves canonical tag by ID
func (r *tagRepository) GetCanonicalByID(ctx context.Context, id uuid.UUID) (*domain.CanonicalTag, error) {
	var canonical domain.CanonicalTag
	err := r.db.WithContext(ctx).
		Preload("Aliases").
		Where("id = ?", id).
		First(&canonical).Error

	if err != nil {
		return nil, fmt.Errorf("canonical tag not found: %w", err)
	}

	return &canonical, nil
}

// UpdateCanonicalTag updates a canonical tag
func (r *tagRepository) UpdateCanonicalTag(ctx context.Context, canonical *domain.CanonicalTag) error {
	return r.db.WithContext(ctx).Save(canonical).Error
}

// ListCanonicalTags returns paginated list of approved canonical tags only
func (r *tagRepository) ListCanonicalTags(ctx context.Context, page, limit int) ([]domain.CanonicalTag, int64, error) {
	var canonicals []domain.CanonicalTag
	var total int64

	// Count total approved tags only
	if err := r.db.WithContext(ctx).Model(&domain.CanonicalTag{}).
		Where("is_approved = ?", true).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Query with pagination - approved tags only
	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).
		Where("is_approved = ?", true).
		Order("display_name ASC").
		Offset(offset).
		Limit(limit).
		Find(&canonicals).Error; err != nil {
		return nil, 0, err
	}

	return canonicals, total, nil
}

// SearchCanonicalTags searches canonical tags using hybrid approach
func (r *tagRepository) SearchCanonicalTags(ctx context.Context, query string, limit int, approvedOnly bool) ([]domain.CanonicalTag, error) {
	var canonicals []domain.CanonicalTag

	// Phase 1: SQL LIKE search (fast & free)
	db := r.db.WithContext(ctx).
		Where("display_name ILIKE ?", "%"+query+"%")

	if approvedOnly {
		db = db.Where("is_approved = ?", true)
	}

	if err := db.Order("display_name ASC").
		Limit(limit).
		Find(&canonicals).Error; err != nil {
		return nil, fmt.Errorf("sql search failed: %w", err)
	}

	// If found results, return immediately
	if len(canonicals) > 0 {
		return canonicals, nil
	}

	// Phase 2: Vector search via aliases
	if r.openAIClient == nil {
		return canonicals, nil // No OpenAI = return empty
	}

	// Generate embedding for query
	embedding, err := r.openAIClient.GetEmbedding(ctx, query)
	if err != nil {
		fmt.Printf("Warning: Vector search failed for query '%s': %v\n", query, err)
		return canonicals, nil
	}

	// Find similar aliases
	type resultRow struct {
		CanonicalTagID uuid.UUID
		Distance       float64
	}
	var results []resultRow

	sqlQuery := `
		SELECT DISTINCT
			canonical_tag_id,
			MIN(embedding <=> $1::vector) as distance
		FROM tag_aliases
		WHERE embedding IS NOT NULL
		GROUP BY canonical_tag_id
		ORDER BY distance ASC
		LIMIT $2
	`

	if err := r.db.WithContext(ctx).Raw(sqlQuery, embedding, limit).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Extract canonical tag IDs
	canonicalIDs := make([]uuid.UUID, len(results))
	for i, r := range results {
		canonicalIDs[i] = r.CanonicalTagID
	}

	// Load canonical tags
	if len(canonicalIDs) > 0 {
		db := r.db.WithContext(ctx).Where("id IN ?", canonicalIDs)

		if approvedOnly {
			db = db.Where("is_approved = ?", true)
		}

		if err := db.Find(&canonicals).Error; err != nil {
			return nil, fmt.Errorf("failed to load canonical tags: %w", err)
		}
	}

	return canonicals, nil
}

// ============================================================
// Video-Canonical Tag Relationship
// ============================================================

// AddCanonicalTagToVideo links a canonical tag to a video
func (r *tagRepository) AddCanonicalTagToVideo(ctx context.Context, videoID, canonicalTagID uuid.UUID) error {
	// Check if relationship already exists
	var count int64
	r.db.WithContext(ctx).
		Table("video_canonical_tags").
		Where("video_id = ? AND canonical_tag_id = ?", videoID, canonicalTagID).
		Count(&count)

	if count > 0 {
		return nil // Already exists
	}

	// Insert relationship
	return r.db.WithContext(ctx).
		Exec("INSERT INTO video_canonical_tags (video_id, canonical_tag_id) VALUES (?, ?)", videoID, canonicalTagID).
		Error
}

// RemoveCanonicalTagFromVideo unlinks a canonical tag from a video
func (r *tagRepository) RemoveCanonicalTagFromVideo(ctx context.Context, videoID, canonicalTagID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Exec("DELETE FROM video_canonical_tags WHERE video_id = ? AND canonical_tag_id = ?", videoID, canonicalTagID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tag not found on video")
	}

	return nil
}

// GetCanonicalTagsByVideoID returns all canonical tags for a video
func (r *tagRepository) GetCanonicalTagsByVideoID(ctx context.Context, videoID uuid.UUID) ([]domain.CanonicalTag, error) {
	var canonicals []domain.CanonicalTag

	err := r.db.WithContext(ctx).Raw(`
		SELECT ct.* 
		FROM canonical_tags ct
		JOIN video_canonical_tags vct ON ct.id = vct.canonical_tag_id
		WHERE vct.video_id = ?
		ORDER BY ct.display_name ASC
	`, videoID).Scan(&canonicals).Error

	if err != nil {
		return nil, err
	}

	return canonicals, nil
}

// ============================================================
// Tag Merge Operations Implementation
// ============================================================

// MergeTags merges source canonical tag into target canonical tag
// Transaction ensures atomicity:
// 1. Move all aliases from source to target
// 2. Update video relationships from source to target
// 3. Delete source canonical tag
func (r *tagRepository) MergeTags(ctx context.Context, sourceID, targetID uuid.UUID) (int, error) {
	var mergedCount int

	// Use transaction to ensure atomicity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Validate both tags exist
		var sourceTag, targetTag domain.CanonicalTag

		if err := tx.First(&sourceTag, "id = ?", sourceID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("source tag not found")
			}
			return fmt.Errorf("failed to find source tag: %w", err)
		}

		if err := tx.First(&targetTag, "id = ?", targetID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("target tag not found")
			}
			return fmt.Errorf("failed to find target tag: %w", err)
		}

		// 2. Move all aliases from source to target
		result := tx.Model(&domain.TagAlias{}).
			Where("canonical_tag_id = ?", sourceID).
			Update("canonical_tag_id", targetID)

		if result.Error != nil {
			return fmt.Errorf("failed to move aliases: %w", result.Error)
		}
		mergedCount = int(result.RowsAffected)

		// 3. Update video relationships: Replace source with target
		// Use raw SQL to handle potential duplicates (ignore conflicts)
		if err := tx.Exec(`
			INSERT INTO video_canonical_tags (video_id, canonical_tag_id)
			SELECT DISTINCT vct.video_id, ?::uuid as canonical_tag_id
			FROM video_canonical_tags vct
			WHERE vct.canonical_tag_id = ?::uuid
			ON CONFLICT (video_id, canonical_tag_id) DO NOTHING
		`, targetID, sourceID).Error; err != nil {
			return fmt.Errorf("failed to update video relationships: %w", err)
		}

		// 4. Delete old video relationships with source tag
		if err := tx.Exec("DELETE FROM video_canonical_tags WHERE canonical_tag_id = ?", sourceID).Error; err != nil {
			return fmt.Errorf("failed to delete old video relationships: %w", err)
		}

		// 5. Delete source canonical tag
		if err := tx.Delete(&sourceTag).Error; err != nil {
			return fmt.Errorf("failed to delete source canonical tag: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return mergedCount, nil
}

// GetAliasCountByCanonicalID returns the number of aliases for a canonical tag
func (r *tagRepository) GetAliasCountByCanonicalID(ctx context.Context, canonicalID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.TagAlias{}).
		Where("canonical_tag_id = ?", canonicalID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}
