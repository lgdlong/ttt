package repository

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
