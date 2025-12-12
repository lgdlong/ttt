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

type TagRepository interface {
	// ============================================================
	// Canonical-Alias Architecture (New)
	// ============================================================

	// GetCanonicalByAlias finds canonical tag by exact normalized text match (Layer 1)
	// Example: "tiền" → CanonicalTag{ID: uuid, DisplayName: "Money"}
	// Returns nil if not found (not an error, proceed to Layer 2)
	GetCanonicalByAlias(ctx context.Context, normalizedText string) (*domain.CanonicalTag, error)

	// GetClosestCanonical finds most similar canonical tag using vector search (Layer 3)
	// Returns (CanonicalTag, similarity_score, error)
	// If no match above threshold, returns (nil, 0, nil)
	GetClosestCanonical(ctx context.Context, embedding pgvector.Vector, threshold float64) (*domain.CanonicalTag, float64, error)

	// CreateCanonicalTag creates a new canonical tag with its initial alias (atomic transaction)
	// Used when: No similar tag found (Layer 4 - Scenario B)
	CreateCanonicalTag(ctx context.Context, canonical *domain.CanonicalTag, initialAlias *domain.TagAlias) error

	// CreateAlias adds a new alias to existing canonical tag
	// Used when: Similar tag found (Layer 4 - Scenario A)
	CreateAlias(ctx context.Context, alias *domain.TagAlias) error

	// GetCanonicalByID retrieves canonical tag by ID
	GetCanonicalByID(ctx context.Context, id uuid.UUID) (*domain.CanonicalTag, error)

	// ListCanonicalTags returns paginated list of canonical tags
	ListCanonicalTags(ctx context.Context, page, limit int) ([]domain.CanonicalTag, int64, error)

	// SearchCanonicalTags searches canonical tags (hybrid: SQL LIKE + Vector)
	SearchCanonicalTags(ctx context.Context, query string, limit int) ([]domain.CanonicalTag, error)

	// ============================================================
	// Video-Canonical Tag Relationship
	// ============================================================

	// AddCanonicalTagToVideo links a canonical tag to a video
	AddCanonicalTagToVideo(ctx context.Context, videoID, canonicalTagID uuid.UUID) error

	// RemoveCanonicalTagFromVideo unlinks a canonical tag from a video
	RemoveCanonicalTagFromVideo(ctx context.Context, videoID, canonicalTagID uuid.UUID) error

	// GetCanonicalTagsByVideoID returns all canonical tags for a video
	GetCanonicalTagsByVideoID(ctx context.Context, videoID uuid.UUID) ([]domain.CanonicalTag, error)

	// ============================================================
	// Tag Merge Operations
	// ============================================================

	// MergeTags merges source canonical tag into target canonical tag
	// All aliases and relationships of source will be moved to target
	// Source canonical tag will be deleted
	// Returns number of aliases merged and any error
	MergeTags(ctx context.Context, sourceID, targetID uuid.UUID) (int, error)

	// GetAliasCountByCanonicalID returns the number of aliases for a canonical tag
	GetAliasCountByCanonicalID(ctx context.Context, canonicalID uuid.UUID) (int, error)

	// ============================================================
	// Translation Layer (New)
	// ============================================================

	// TranslateText translates text to English using OpenAI GPT-4o-mini
	// Used to normalize cross-lingual queries before vector search
	// Returns empty string if OpenAI client is not available
	TranslateText(ctx context.Context, text string) (string, error)

	// ============================================================
	// Shared Utilities
	// ============================================================
	GetEmbeddingForText(ctx context.Context, text string) ([]float32, error)

	// ============================================================
	// Legacy Tag CRUD (DEPRECATED - Removed, use Tag V2 API)
	// ============================================================
	// Removed all legacy Tag methods to enforce use of CanonicalTag + TagAlias architecture
}

type tagRepository struct {
	db           *gorm.DB
	openAIClient *infrastructure.OpenAIClient
}

func NewTagRepository(db *gorm.DB, openAIClient *infrastructure.OpenAIClient) TagRepository {
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
// Legacy Tag CRUD Implementation (DEPRECATED - Commented Out)
// ============================================================
/*
// Create creates a new tag với auto-generate embedding
func (r *tagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	// CRITICAL: Luôn generate embedding khi tạo tag mới
	// Nếu không có embedding, vector search sẽ không tìm thấy tag này
	if r.openAIClient != nil {
		embedding, err := r.openAIClient.GetEmbedding(ctx, tag.Name)
		if err != nil {
			// Log warning but don't fail - tag vẫn có thể tạo được
			// Chỉ mất khả năng semantic search
			fmt.Printf("Warning: Failed to generate embedding for tag '%s': %v\n", tag.Name, err)
		} else {
			tag.Embedding = embedding
		}
	}

	return r.db.Create(tag).Error
}

// GetByID retrieves a tag by ID
func (r *tagRepository) GetByID(id uuid.UUID) (*domain.Tag, error) {
	var tag domain.Tag
	if err := r.db.First(&tag, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByName retrieves a tag by name
func (r *tagRepository) GetByName(name string) (*domain.Tag, error) {
	var tag domain.Tag
	if err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// Update updates a tag
func (r *tagRepository) Update(tag *domain.Tag) error {
	return r.db.Save(tag).Error
}

// Delete deletes a tag
func (r *tagRepository) Delete(id uuid.UUID) error {
	// Delete from video_tags first (cascade)
	if err := r.db.Exec("DELETE FROM video_tags WHERE tag_id = ?", id).Error; err != nil {
		return err
	}
	return r.db.Delete(&domain.Tag{}, "id = ?", id).Error
}

// List returns paginated list of tags
func (r *tagRepository) List(page, limit int) ([]domain.Tag, int64, error) {
	var tags []domain.Tag
	var total int64

	if err := r.db.Model(&domain.Tag{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.Order("name ASC").Offset(offset).Limit(limit).Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// Search implements HYBRID SEARCH strategy:
// 1. SQL LIKE first (fast, free, exact matching)
// 2. Vector Search if no results (smart, costs money, semantic matching)
func (r *tagRepository) Search(ctx context.Context, query string, limit int) ([]domain.Tag, error) {
	var tags []domain.Tag

	// --- PHASE 1: SQL LIKE Search (Ưu tiên) ---
	// Tìm theo tên chính xác hoặc gần đúng
	// ILIKE = case-insensitive LIKE trong PostgreSQL
	if err := r.db.Where("name ILIKE ?", "%"+query+"%").
		Order("name ASC").
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("sql search failed: %w", err)
	}

	// [Decision Point] Đã tìm thấy kết quả? → Return ngay!
	// => Chi phí OpenAI: $0
	if len(tags) > 0 {
		return tags, nil
	}

	// --- PHASE 2: Vector Search (Fallback khi SQL thất bại) ---
	// Chỉ chạy đến đây nếu:
	// - SQL không tìm thấy gì
	// - OpenAI client available
	if r.openAIClient == nil {
		// Không có OpenAI client → trả về empty
		return tags, nil
	}

	// 1. Convert query sang embedding vector (Tốn tiền ở đây!)
	queryVector, err := r.openAIClient.GetEmbedding(ctx, query)
	if err != nil {
		// Nếu OpenAI API fail (hết tiền, network error, rate limit)
		// → Không crash app, chỉ log warning và return empty
		fmt.Printf("Warning: Vector search failed for query '%s': %v\n", query, err)
		return tags, nil
	}

	// 2. Tìm trong DB bằng cosine distance
	// Operator <=> là cosine distance trong pgvector
	// ORDER BY ... ASC → khoảng cách nhỏ nhất = giống nhất
	if err := r.db.Model(&domain.Tag{}).
		Order(gorm.Expr("embedding <=> ?", queryVector)).
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	return tags, nil
}

// AddTagToVideo adds a tag to a video
func (r *tagRepository) AddTagToVideo(videoID, tagID uuid.UUID) error {
	// Check if relationship already exists
	var count int64
	r.db.Table("video_tags").Where("video_id = ? AND tag_id = ?", videoID, tagID).Count(&count)
	if count > 0 {
		return nil // Already exists
	}

	return r.db.Exec("INSERT INTO video_tags (video_id, tag_id) VALUES (?, ?)", videoID, tagID).Error
}

// RemoveTagFromVideo removes a tag from a video
func (r *tagRepository) RemoveTagFromVideo(videoID, tagID uuid.UUID) error {
	result := r.db.Exec("DELETE FROM video_tags WHERE video_id = ? AND tag_id = ?", videoID, tagID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tag not found on video")
	}
	return nil
}

// GetTagsByVideoID returns all tags for a video
func (r *tagRepository) GetTagsByVideoID(videoID uuid.UUID) ([]domain.Tag, error) {
	var tags []domain.Tag
	if err := r.db.Raw(`
		SELECT t.* FROM tags t
		JOIN video_tags vt ON t.id = vt.tag_id
		WHERE vt.video_id = ?
		ORDER BY t.name ASC
	`, videoID).Scan(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
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
	// pgvector.Vector is internally a []float32, we can access it via slice conversion
	vectorSlice := make([]float32, len(embedding.Slice()))
	copy(vectorSlice, embedding.Slice())

	return vectorSlice, nil
}

// FindSimilarTag finds the most similar tag using cosine distance
// Returns the tag, its distance, and error
// If no tag found within threshold, returns nil tag with distance 1.0

// FindSimilarTag finds the most similar tag using cosine distance
func (r *tagRepository) FindSimilarTag(ctx context.Context, embedding []float32, threshold float64) (*domain.Tag, float64, error) {
	tags, distances, err := r.FindTopSimilarTags(ctx, embedding, threshold, 1)
	if err != nil || len(tags) == 0 {
		return nil, 1.0, err
	}
	return &tags[0], distances[0], nil
}

// FindTopSimilarTags finds the top N most similar tags within a threshold
func (r *tagRepository) FindTopSimilarTags(ctx context.Context, embedding []float32, threshold float64, limit int) ([]domain.Tag, []float64, error) {
	if r.openAIClient == nil {
		fmt.Printf("[REPO] OpenAI client is nil\n")
		return nil, nil, nil // No similarity check if OpenAI not available
	}

	fmt.Printf("[REPO] FindTopSimilarTags: threshold=%.2f, limit=%d\n", threshold, limit)

	// First, check how many tags with embeddings exist
	var countWithEmbedding int64
	r.db.Model(&domain.Tag{}).Where("embedding IS NOT NULL").Count(&countWithEmbedding)
	fmt.Printf("[REPO] Tags with embeddings in DB: %d\n", countWithEmbedding)

	type resultRow struct {
		ID        uuid.UUID
		Name      string
		Embedding pgvector.Vector // Changed from []float32 to pgvector.Vector
		Distance  float64
	}
	var results []resultRow

	queryVec := pgvector.NewVector(embedding)
	fmt.Printf("[REPO] Query vector created (size: %d)\n", len(embedding))
	fmt.Printf("[REPO] Input params: threshold=%.4f, limit=%d\n", threshold, limit)

	// Use pgx parameter binding properly
	// Note: LIMIT doesn't accept ? parameter in pgx, must pass as part of query
	sqlQuery := fmt.Sprintf(`
		SELECT
			id, name, embedding,
			embedding <=> $1::vector as distance
		FROM tags
		WHERE embedding IS NOT NULL
			AND embedding <=> $1::vector < $2
		ORDER BY embedding <=> $1::vector ASC
		LIMIT %d
	`, limit)

	fmt.Printf("[REPO] SQL Query:\n%s\n", sqlQuery)
	fmt.Printf("[REPO] Executing with params: queryVec, threshold=%.4f\n", threshold)

	err := r.db.Raw(sqlQuery, queryVec, threshold).Scan(&results).Error

	if err != nil {
		fmt.Printf("[REPO] ✗ SQL Error: %v\n", err)
		return nil, nil, fmt.Errorf("failed to query similar tags: %w", err)
	}

	fmt.Printf("[REPO] ✓ Query returned %d results\n", len(results))

	tags := make([]domain.Tag, 0, len(results))
	distances := make([]float64, 0, len(results))
	for i, row := range results {
		similarity := (1.0 - row.Distance/2.0) * 100
		fmt.Printf("[REPO]   [%d] '%s' (distance: %.4f, similarity: %.1f%%)\n", i+1, row.Name, row.Distance, similarity)
		tags = append(tags, domain.Tag{
			ID:        row.ID,
			Name:      row.Name,
			Embedding: row.Embedding,
		})
		distances = append(distances, row.Distance)
	}
	return tags, distances, nil
}

// GetOrCreateByName gets a tag by name or creates it if not exists
func (r *tagRepository) GetOrCreateByName(ctx context.Context, name string) (*domain.Tag, error) {
	tag, err := r.GetByName(name)
	if err == nil {
		return tag, nil
	}

	// Create new tag with embedding
	newTag := &domain.Tag{
		Name: name,
	}
	if err := r.Create(ctx, newTag); err != nil {
		return nil, err
	}
	return newTag, nil
}
*/

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

// ListCanonicalTags returns paginated list of canonical tags
func (r *tagRepository) ListCanonicalTags(ctx context.Context, page, limit int) ([]domain.CanonicalTag, int64, error) {
	var canonicals []domain.CanonicalTag
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&domain.CanonicalTag{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Query with pagination
	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).
		Order("display_name ASC").
		Offset(offset).
		Limit(limit).
		Find(&canonicals).Error; err != nil {
		return nil, 0, err
	}

	return canonicals, total, nil
}

// SearchCanonicalTags searches canonical tags using hybrid approach
func (r *tagRepository) SearchCanonicalTags(ctx context.Context, query string, limit int) ([]domain.CanonicalTag, error) {
	var canonicals []domain.CanonicalTag

	// Phase 1: SQL LIKE search (fast & free)
	if err := r.db.WithContext(ctx).
		Where("display_name ILIKE ?", "%"+query+"%").
		Order("display_name ASC").
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
		if err := r.db.WithContext(ctx).
			Where("id IN ?", canonicalIDs).
			Find(&canonicals).Error; err != nil {
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
