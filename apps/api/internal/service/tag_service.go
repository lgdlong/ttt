package service

import (
	"api/internal/repository"
)

// SemanticDuplicateError - REMOVED (legacy Tag V1 struct)

type TagService interface {
	// ============================================================
	// Legacy Tag CRUD - REMOVED (use Tag V2 API)
	// ============================================================
	// All legacy Tag methods removed to enforce CanonicalTag + TagAlias architecture
}

type tagService struct {
	tagRepo   repository.TagRepository
	videoRepo repository.VideoRepository
}

func NewTagService(tagRepo repository.TagRepository, videoRepo repository.VideoRepository) TagService {
	return &tagService{
		tagRepo:   tagRepo,
		videoRepo: videoRepo,
	}
}

/*
// ============================================================
// Legacy Tag Service Methods - REMOVED (use Tag V2 API)
// ============================================================

// CreateTag creates a new tag with semantic deduplication
// Flow: 1) Generate embedding → 2) Check similarity → 3) Create or return duplicate error
func (s *tagService) CreateTag(ctx context.Context, req dto.CreateTagRequest) (*dto.TagResponse, error) {
	// Step 1: Check exact name match first (cheap operation)
	fmt.Printf("[TAG_SERVICE] Step 1: Checking exact name match for '%s'...\n", req.Name)
	if existing, _ := s.tagRepo.GetByName(req.Name); existing != nil {
		errMsg := fmt.Sprintf("tag with name '%s' already exists", req.Name)
		fmt.Printf("[TAG_SERVICE] ✗ Exact match found: %s\n", errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	fmt.Printf("[TAG_SERVICE] ✓ No exact name match\n")

	// Step 2: Generate embedding for the new tag name
	fmt.Printf("[TAG_SERVICE] Step 2: Generating embedding...\n")
	embedding, err := s.tagRepo.GetEmbeddingForText(ctx, req.Name)
	if err != nil {
		fmt.Printf("[TAG_SERVICE] ⚠ OpenAI unavailable: %v, creating tag without semantic check\n", err)
		// If OpenAI unavailable, proceed without semantic check
		tag := &domain.Tag{Name: req.Name}
		if createErr := s.tagRepo.Create(ctx, tag); createErr != nil {
			return nil, fmt.Errorf("failed to create tag: %w", createErr)
		}
		return s.toTagResponse(tag), nil
	}
	fmt.Printf("[TAG_SERVICE] ✓ Embedding generated (size: %d dimensions)\n", len(embedding))

	// Step 3: Check for semantically similar tags
	// Distance range [0, 2]: 0=identical, 1=orthogonal, 2=opposite
	// Threshold 0.30 = 85% semantic similarity (strict)
	// This catches cross-language synonyms (e.g., "money" ↔ "tiền", "video" ↔ "clip")
	// But REJECTS unrelated words (e.g., "money" vs "sex")
	const SIMILARITY_THRESHOLD = 0.30
	const SUGGESTION_LIMIT = 3 // Return top 3 similar tags as suggestions
	fmt.Printf("[TAG_SERVICE] Step 3: Searching for similar tags (threshold: %.2f, limit: %d)...\n", SIMILARITY_THRESHOLD, SUGGESTION_LIMIT)
	similarTags, distances, err := s.tagRepo.FindTopSimilarTags(ctx, embedding, SIMILARITY_THRESHOLD, SUGGESTION_LIMIT)
	if err != nil {
		fmt.Printf("[TAG_SERVICE] ✗ Error checking similarity: %v\n", err)
		return nil, fmt.Errorf("failed to check similarity: %w", err)
	}

	// Step 4: If similar tag found, return semantic duplicate error with suggestions
	if len(similarTags) > 0 {
		fmt.Printf("[TAG_SERVICE] ⚠ Found %d similar tags:\n", len(similarTags))
		for i, tag := range similarTags {
			similarity := (1.0 - distances[i]/2.0) * 100
			fmt.Printf("  [%d] '%s' (distance: %.4f, similarity: %.1f%%)\n", i+1, tag.Name, distances[i], similarity)
		}
		return nil, &SemanticDuplicateError{
			ExistingTag: &similarTags[0],
			Distance:    distances[0],
			Suggestions: similarTags,
		}
	}
	fmt.Printf("[TAG_SERVICE] ✓ No similar tags found\n")

	// Step 5: No duplicate found - proceed to create tag with embedding
	fmt.Printf("[TAG_SERVICE] Step 4: Creating tag...\n")
	tag := &domain.Tag{
		Name: req.Name,
	}

	// Note: Create() will auto-generate embedding again, but we could optimize
	// by passing the embedding we already generated
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		fmt.Printf("[TAG_SERVICE] ✗ Failed to create tag: %v\n", err)
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	fmt.Printf("[TAG_SERVICE] ✓ Tag created successfully (ID: %s)\n", tag.ID.String())
	return s.toTagResponse(tag), nil
}

// GetTagByID retrieves a tag by ID
func (s *tagService) GetTagByID(id string) (*dto.TagResponse, error) {
	tagUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid tag ID: %w", err)
	}

	tag, err := s.tagRepo.GetByID(tagUUID)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	return s.toTagResponse(tag), nil
}

// UpdateTag updates a tag
func (s *tagService) UpdateTag(id string, req dto.UpdateTagRequest) (*dto.TagResponse, error) {
	tagUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid tag ID: %w", err)
	}

	tag, err := s.tagRepo.GetByID(tagUUID)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// Check if new name conflicts with existing tag
	if existing, _ := s.tagRepo.GetByName(req.Name); existing != nil && existing.ID != tag.ID {
		return nil, fmt.Errorf("tag with name '%s' already exists", req.Name)
	}

	tag.Name = req.Name
	if err := s.tagRepo.Update(tag); err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return s.toTagResponse(tag), nil
}

// DeleteTag deletes a tag
func (s *tagService) DeleteTag(id string) error {
	tagUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid tag ID: %w", err)
	}

	if _, err := s.tagRepo.GetByID(tagUUID); err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	if err := s.tagRepo.Delete(tagUUID); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// ListTags returns paginated list of tags
func (s *tagService) ListTags(ctx context.Context, req dto.TagListRequest) (*dto.TagListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// If query provided, search instead of list (Hybrid Search: SQL + Vector)
	if req.Query != "" {
		tags, err := s.tagRepo.Search(ctx, req.Query, req.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search tags: %w", err)
		}

		tagResponses := make([]dto.TagResponse, len(tags))
		for i, tag := range tags {
			tagResponses[i] = *s.toTagResponse(&tag)
		}

		return &dto.TagListResponse{
			Data: tagResponses,
			Pagination: dto.PaginationMetadata{
				Page:       1,
				Limit:      req.Limit,
				TotalItems: int64(len(tags)),
				TotalPages: 1,
			},
		}, nil
	}

	tags, total, err := s.tagRepo.List(req.Page, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	tagResponses := make([]dto.TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = *s.toTagResponse(&tag)
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &dto.TagListResponse{
		Data: tagResponses,
		Pagination: dto.PaginationMetadata{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

// SearchTags searches tags using hybrid search (SQL LIKE → Vector Search)
func (s *tagService) SearchTags(ctx context.Context, query string, limit int) ([]dto.TagResponse, error) {
	if limit < 1 {
		limit = 20
	}

	tags, err := s.tagRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search tags: %w", err)
	}

	tagResponses := make([]dto.TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = *s.toTagResponse(&tag)
	}

	return tagResponses, nil
}

// AddTagToVideo adds a tag to a video (creates tag if not exists)
func (s *tagService) AddTagToVideo(ctx context.Context, videoID string, req dto.AddVideoTagRequest) (*dto.TagResponse, error) {
	videoUUID, err := uuid.Parse(videoID)
	if err != nil {
		return nil, fmt.Errorf("invalid video ID: %w", err)
	}

	// Verify video exists
	if _, err := s.videoRepo.GetVideoByID(videoUUID); err != nil {
		return nil, fmt.Errorf("video not found: %w", err)
	}

	var tag *domain.Tag

	// If tag_id provided, use existing tag
	if req.TagID != nil && *req.TagID != "" {
		tagUUID, err := uuid.Parse(*req.TagID)
		if err != nil {
			return nil, fmt.Errorf("invalid tag ID: %w", err)
		}
		tag, err = s.tagRepo.GetByID(tagUUID)
		if err != nil {
			return nil, fmt.Errorf("tag not found: %w", err)
		}
	} else if req.TagName != nil && *req.TagName != "" {
		// Get or create tag by name (with embedding)
		tag, err = s.tagRepo.GetOrCreateByName(ctx, *req.TagName)
		if err != nil {
			return nil, fmt.Errorf("failed to get or create tag: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either tag_id or tag_name must be provided")
	}

	// Add tag to video
	if err := s.tagRepo.AddTagToVideo(videoUUID, tag.ID); err != nil {
		return nil, fmt.Errorf("failed to add tag to video: %w", err)
	}

	return s.toTagResponse(tag), nil
}

// RemoveTagFromVideo removes a tag from a video
func (s *tagService) RemoveTagFromVideo(videoID, tagID string) error {
	videoUUID, err := uuid.Parse(videoID)
	if err != nil {
		return fmt.Errorf("invalid video ID: %w", err)
	}

	tagUUID, err := uuid.Parse(tagID)
	if err != nil {
		return fmt.Errorf("invalid tag ID: %w", err)
	}

	if err := s.tagRepo.RemoveTagFromVideo(videoUUID, tagUUID); err != nil {
		return fmt.Errorf("failed to remove tag from video: %w", err)
	}

	return nil
}

// GetVideoTags returns all tags for a video
func (s *tagService) GetVideoTags(videoID string) ([]dto.TagResponse, error) {
	videoUUID, err := uuid.Parse(videoID)
	if err != nil {
		return nil, fmt.Errorf("invalid video ID: %w", err)
	}

	tags, err := s.tagRepo.GetTagsByVideoID(videoUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video tags: %w", err)
	}

	tagResponses := make([]dto.TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = *s.toTagResponse(&tag)
	}

	return tagResponses, nil
}

// Helper: Convert domain.Tag to dto.TagResponse
func (s *tagService) toTagResponse(tag *domain.Tag) *dto.TagResponse {
	return &dto.TagResponse{
		ID:   tag.ID.String(),
		Name: tag.Name,
	}
}
*/
