package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type tagServiceV2 struct {
	tagRepo   domain.TagRepository
	videoRepo domain.VideoRepository
}

// NewTagServiceV2 creates a new v2 tag service instance
func NewTagServiceV2(tagRepo domain.TagRepository, videoRepo domain.VideoRepository) domain.TagServiceV2 {
	return &tagServiceV2{
		tagRepo:   tagRepo,
		videoRepo: videoRepo,
	}
}

// ============================================================
// Canonical-Alias Architecture Implementation
// ============================================================

// AUTO_MERGE_THRESHOLD controls when to auto-merge similar tags
// This is a DISTANCE threshold (cosine distance in pgvector)
// Distance formula: similarity = 1 - (distance / 2)
//
// Distance 0.20 = 90% similarity (very strict)
// Distance 0.30 = 85% similarity (recommended)
// Distance 0.40 = 80% similarity (too loose)
//
// Examples with threshold 0.30:
// ✅ "money" ↔ "tiền" (distance ~0.25) → MERGE
// ✅ "ML" ↔ "Machine Learning" (distance ~0.20) → MERGE
// ❌ "money" ↔ "gái" (distance ~0.80) → DON'T MERGE
// ❌ "finance" ↔ "sex" (distance ~1.50) → DON'T MERGE
const AUTO_MERGE_THRESHOLD = 0.40 // 85% similarity

// ResolveTag implements 4-layer resolution algorithm to get or create canonical tag
// This is the CORE method that replaces CreateTag with zero-error flow
//
// Algorithm:
// Layer 1: Exact Match (Cache Hit) - Fast & Free
// Layer 2: Embedding Generation - Required for Layer 3
// Layer 3: Semantic Search - AI-powered similarity
// Layer 4: Decision Making - Auto-merge or create new
//
// Returns: (canonical, matchedAlias, isNewTag, error)
// - canonical: The resolved canonical tag
// - matchedAlias: The alias text that was matched (for UI feedback)
// - isNewTag: true if new canonical was created, false if existing
// - error: Only critical errors (DB failures, etc.), NOT duplicate errors
func (s *tagServiceV2) ResolveTag(ctx context.Context, userInput string) (*domain.CanonicalTag, string, bool, error) {
	fmt.Printf("\n[RESOLVE_TAG] ========================================\n")
	fmt.Printf("[RESOLVE_TAG] Input: '%s'\n", userInput)

	// ============================================================
	// Layer 1: Exact Match (Normalized Text Lookup)
	// Cost: ~10ms, $0
	// ============================================================
	fmt.Printf("[RESOLVE_TAG] Layer 1: Exact Match Check...\n")
	normalizedInput := domain.NormalizeText(userInput)
	fmt.Printf("[RESOLVE_TAG]   Normalized: '%s'\n", normalizedInput)

	canonical, err := s.tagRepo.GetCanonicalByAlias(ctx, normalizedInput)
	if err != nil {
		return nil, "", false, fmt.Errorf("Layer 1 failed: %w", err)
	}

	if canonical != nil {
		fmt.Printf("[RESOLVE_TAG] ✓ Layer 1 HIT: Found canonical '%s' (ID: %s)\n", canonical.DisplayName, canonical.ID)
		fmt.Printf("[RESOLVE_TAG] ========================================\n\n")
		return canonical, normalizedInput, false, nil
	}

	fmt.Printf("[RESOLVE_TAG] ✗ Layer 1 MISS: No exact match found\n")

	// ============================================================
	// Layer 1.5: Translation Layer (Cross-Lingual Support)
	// Cost: ~300ms, ~$0.00015 (GPT-4o-mini API call)
	// Purpose: Normalize non-English queries to English before vector search
	// Example: "Tiền" -> "Money", "Học máy" -> "Machine Learning"
	// ============================================================
	fmt.Printf("[RESOLVE_TAG] Layer 1.5: Translation Layer...\n")

	// Attempt translation to English
	englishTerm, err := s.tagRepo.TranslateText(ctx, userInput)

	// Process translation result only if successful and different from original
	if err == nil && englishTerm != "" && domain.NormalizeText(englishTerm) != normalizedInput {
		fmt.Printf("[RESOLVE_TAG]   Translated: '%s' -> '%s'\n", userInput, englishTerm)

		// Check if translated English term has existing canonical tag
		normalizedEng := domain.NormalizeText(englishTerm)
		canonicalEng, err := s.tagRepo.GetCanonicalByAlias(ctx, normalizedEng)

		if err == nil && canonicalEng != nil {
			fmt.Printf("[RESOLVE_TAG] ✓ Layer 1.5 HIT: Found canonical via translation '%s'\n", canonicalEng.DisplayName)

			// Auto-create alias for original input to avoid future translation costs
			// Use original input's embedding (not translated term) for future semantic search
			embeddingSlice, embErr := s.tagRepo.GetEmbeddingForText(ctx, userInput)
			if embErr == nil {
				embedding := pgvector.NewVector(embeddingSlice)
				newAlias, aliasErr := domain.NewTagAlias(userInput, canonicalEng.ID, embedding, 1.0)
				if aliasErr != nil {
					fmt.Printf("[RESOLVE_TAG] ⚠ Warning: Failed to create alias: %v\n", aliasErr)
				} else {
					// Save alias to optimize future lookups (next time will hit Layer 1)
					if createErr := s.tagRepo.CreateAlias(ctx, newAlias); createErr != nil {
						// Log warning but don't fail request - still return found canonical
						fmt.Printf("[RESOLVE_TAG] ⚠ Warning: Failed to save translation alias: %v\n", createErr)
					} else {
						fmt.Printf("[RESOLVE_TAG] ✓ Created translation alias '%s' -> '%s'\n", userInput, canonicalEng.DisplayName)
					}
				}
			}

			fmt.Printf("[RESOLVE_TAG] ========================================\n\n")
			return canonicalEng, userInput, false, nil
		}

		fmt.Printf("[RESOLVE_TAG] ✗ Layer 1.5 MISS: Translated term '%s' not found in DB\n", englishTerm)
	} else if err != nil {
		fmt.Printf("[RESOLVE_TAG] ⚠ Translation failed: %v (continuing to Layer 2)\n", err)
	} else {
		fmt.Printf("[RESOLVE_TAG]   Skipped: Input already in English or translation returned same term\n")
	}

	// ============================================================
	// Layer 2: Embedding Generation
	// Cost: ~500ms, ~$0.0001 (OpenAI API call)
	// ============================================================
	fmt.Printf("[RESOLVE_TAG] Layer 2: Generating embedding...\n")

	embeddingSlice, err := s.tagRepo.GetEmbeddingForText(ctx, userInput)
	if err != nil {
		// OpenAI unavailable → Create new canonical without semantic check
		fmt.Printf("[RESOLVE_TAG] ⚠ Layer 2 FAILED: OpenAI unavailable (%v)\n", err)
		fmt.Printf("[RESOLVE_TAG]   Fallback: Creating new canonical without semantic check\n")

		newCanonical, canonicalErr := domain.NewCanonicalTag(userInput)
		if canonicalErr != nil {
			return nil, "", false, fmt.Errorf("failed to create canonical (invalid input): %w", canonicalErr)
		}
		// Note: Using empty embedding since OpenAI failed. This will be backfilled later.
		newAlias, aliasErr := domain.NewInitialTagAlias(userInput, pgvector.Vector{}, 1.0)
		if aliasErr != nil {
			return nil, "", false, fmt.Errorf("failed to create alias (validation failed): %w", aliasErr)
		}

		if createErr := s.tagRepo.CreateCanonicalTag(ctx, newCanonical, newAlias); createErr != nil {
			return nil, "", false, fmt.Errorf("failed to create canonical (no OpenAI): %w", createErr)
		}

		fmt.Printf("[RESOLVE_TAG] ✓ Created new canonical '%s' (ID: %s)\n", newCanonical.DisplayName, newCanonical.ID)
		fmt.Printf("[RESOLVE_TAG] ========================================\n\n")
		return newCanonical, userInput, true, nil
	}

	embedding := pgvector.NewVector(embeddingSlice)
	fmt.Printf("[RESOLVE_TAG] ✓ Layer 2 SUCCESS: Embedding generated (%d dims)\n", len(embeddingSlice))

	// ============================================================
	// Layer 3: Semantic Search (Vector Similarity)
	// Cost: ~50ms, $0 (uses cached embeddings in DB)
	// ============================================================
	fmt.Printf("[RESOLVE_TAG] Layer 3: Semantic search (threshold: %.2f)...\n", AUTO_MERGE_THRESHOLD)

	closestCanonical, similarityScore, err := s.tagRepo.GetClosestCanonical(ctx, embedding, AUTO_MERGE_THRESHOLD)
	if err != nil {
		return nil, "", false, fmt.Errorf("Layer 3 failed: %w", err)
	}

	// ============================================================
	// Layer 4: Decision Making
	// ============================================================

	// Scenario A: Match Found (Score >= Threshold)
	// Action: Create new alias → Link to existing canonical
	if closestCanonical != nil {
		fmt.Printf("[RESOLVE_TAG] ✓ Layer 3 HIT: Found similar canonical '%s' (score: %.2f%%)\n",
			closestCanonical.DisplayName, similarityScore*100)

		fmt.Printf("[RESOLVE_TAG] Layer 4: AUTO-MERGE (Scenario A)\n")
		fmt.Printf("[RESOLVE_TAG]   Action: Create alias '%s' → Canonical '%s'\n",
			userInput, closestCanonical.DisplayName)

		// Create new alias pointing to existing canonical
		newAlias, aliasErr := domain.NewTagAlias(userInput, closestCanonical.ID, embedding, similarityScore)
		if aliasErr != nil {
			return nil, "", false, fmt.Errorf("failed to create alias (validation failed): %w", aliasErr)
		}
		if err := s.tagRepo.CreateAlias(ctx, newAlias); err != nil {
			return nil, "", false, fmt.Errorf("failed to create alias: %w", err)
		}

		fmt.Printf("[RESOLVE_TAG] ✓ Alias created successfully\n")
		fmt.Printf("[RESOLVE_TAG] ========================================\n\n")
		return closestCanonical, userInput, false, nil
	}

	// Scenario B: No Match (Score < Threshold)
	// Action: Create new canonical + initial alias
	fmt.Printf("[RESOLVE_TAG] ✗ Layer 3 MISS: No similar canonical found\n")
	fmt.Printf("[RESOLVE_TAG] Layer 4: CREATE NEW (Scenario B)\n")
	fmt.Printf("[RESOLVE_TAG]   Action: Create new canonical '%s' + initial alias\n", userInput)

	newCanonical, canonicalErr := domain.NewCanonicalTag(userInput)
	if canonicalErr != nil {
		return nil, "", false, fmt.Errorf("failed to create canonical (invalid input): %w", canonicalErr)
	}
	newAlias, aliasErr := domain.NewInitialTagAlias(userInput, embedding, 1.0) // Score=1.0 for canonical
	if aliasErr != nil {
		return nil, "", false, fmt.Errorf("failed to create alias (validation failed): %w", aliasErr)
	}

	if err := s.tagRepo.CreateCanonicalTag(ctx, newCanonical, newAlias); err != nil {
		return nil, "", false, fmt.Errorf("failed to create canonical: %w", err)
	}

	fmt.Printf("[RESOLVE_TAG] ✓ New canonical created (ID: %s)\n", newCanonical.ID)
	fmt.Printf("[RESOLVE_TAG] ========================================\n\n")
	return newCanonical, userInput, true, nil
}

// ============================================================
// Canonical Tag CRUD
// ============================================================

func (s *tagServiceV2) GetCanonicalTagByID(ctx context.Context, id string) (*dto.TagResponse, error) {
	canonicalUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid canonical tag ID: %w", err)
	}

	canonical, err := s.tagRepo.GetCanonicalByID(ctx, canonicalUUID)
	if err != nil {
		return nil, fmt.Errorf("canonical tag not found: %w", err)
	}

	return s.toCanonicalTagResponse(canonical), nil
}

func (s *tagServiceV2) ListCanonicalTags(ctx context.Context, req dto.TagListRequest) (*dto.TagListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// If query provided, search instead of list
	if req.Query != "" {
		canonicals, err := s.tagRepo.SearchCanonicalTags(ctx, req.Query, req.Limit, req.ApprovedOnly)
		if err != nil {
			return nil, fmt.Errorf("failed to search canonical tags: %w", err)
		}

		tagResponses := make([]dto.TagResponse, len(canonicals))
		for i, canonical := range canonicals {
			tagResponses[i] = *s.toCanonicalTagResponse(&canonical)
		}

		return &dto.TagListResponse{
			Data: tagResponses,
			Pagination: dto.PaginationMetadata{
				Page:       1,
				Limit:      req.Limit,
				TotalItems: int64(len(canonicals)),
				TotalPages: 1,
			},
		}, nil
	}

	canonicals, total, err := s.tagRepo.ListCanonicalTags(ctx, req.Page, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list canonical tags: %w", err)
	}

	tagResponses := make([]dto.TagResponse, len(canonicals))
	for i, canonical := range canonicals {
		tagResponses[i] = *s.toCanonicalTagResponse(&canonical)
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

func (s *tagServiceV2) SearchCanonicalTags(ctx context.Context, query string, limit int, approvedOnly bool) ([]dto.TagResponse, error) {
	if limit < 1 {
		limit = 20
	}

	canonicals, err := s.tagRepo.SearchCanonicalTags(ctx, query, limit, approvedOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to search canonical tags: %w", err)
	}

	tagResponses := make([]dto.TagResponse, len(canonicals))
	for i, canonical := range canonicals {
		tagResponses[i] = *s.toCanonicalTagResponse(&canonical)
	}

	return tagResponses, nil
}

// ============================================================
// Video-Canonical Tag Management
// ============================================================

func (s *tagServiceV2) AddCanonicalTagToVideo(ctx context.Context, videoID string, req dto.AddVideoTagRequest) (*dto.TagResponse, error) {
	videoUUID, err := uuid.Parse(videoID)
	if err != nil {
		return nil, fmt.Errorf("invalid video ID: %w", err)
	}

	// Verify video exists
	if _, err := s.videoRepo.GetVideoByID(videoUUID); err != nil {
		return nil, fmt.Errorf("video not found: %w", err)
	}

	var canonical *domain.CanonicalTag

	// If tag_id provided, get existing canonical
	if req.TagID != nil && *req.TagID != "" {
		tagUUID, err := uuid.Parse(*req.TagID)
		if err != nil {
			return nil, fmt.Errorf("invalid tag_id format: %w", err)
		}
		canonical, err = s.tagRepo.GetCanonicalByID(ctx, tagUUID)
		if err != nil {
			return nil, fmt.Errorf("canonical tag not found: %w", err)
		}
	} else if req.TagName != nil && *req.TagName != "" {
		// If tag_name provided, resolve using 4-layer algorithm
		canonical, _, _, err = s.ResolveTag(ctx, *req.TagName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve tag: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either tag_id or tag_name must be provided")
	}

	// Add canonical tag to video
	if err := s.tagRepo.AddCanonicalTagToVideo(ctx, videoUUID, canonical.ID); err != nil {
		return nil, fmt.Errorf("failed to add canonical tag to video: %w", err)
	}

	return s.toCanonicalTagResponse(canonical), nil
}

func (s *tagServiceV2) RemoveCanonicalTagFromVideo(ctx context.Context, videoID, canonicalTagID string) error {
	videoUUID, err := uuid.Parse(videoID)
	if err != nil {
		return fmt.Errorf("invalid video ID: %w", err)
	}

	canonicalUUID, err := uuid.Parse(canonicalTagID)
	if err != nil {
		return fmt.Errorf("invalid canonical tag ID: %w", err)
	}

	if err := s.tagRepo.RemoveCanonicalTagFromVideo(ctx, videoUUID, canonicalUUID); err != nil {
		return fmt.Errorf("failed to remove canonical tag from video: %w", err)
	}

	return nil
}

func (s *tagServiceV2) GetVideoCanonicalTags(ctx context.Context, videoID string) ([]dto.TagResponse, error) {
	videoUUID, err := uuid.Parse(videoID)
	if err != nil {
		return nil, fmt.Errorf("invalid video ID: %w", err)
	}

	canonicals, err := s.tagRepo.GetCanonicalTagsByVideoID(ctx, videoUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video canonical tags: %w", err)
	}

	tagResponses := make([]dto.TagResponse, len(canonicals))
	for i, canonical := range canonicals {
		tagResponses[i] = *s.toCanonicalTagResponse(&canonical)
	}

	return tagResponses, nil
}

// ============================================================
// Helper Methods
// ============================================================

// toCanonicalTagResponse converts domain.CanonicalTag to dto.TagResponse
func (s *tagServiceV2) toCanonicalTagResponse(canonical *domain.CanonicalTag) *dto.TagResponse {
	// Map aliases to string array (RawText field)
	aliases := make([]string, 0, len(canonical.Aliases))
	for _, alias := range canonical.Aliases {
		// Skip if alias matches the canonical name to avoid duplication
		if alias.RawText != canonical.DisplayName {
			aliases = append(aliases, alias.RawText)
		}
	}

	return &dto.TagResponse{
		ID:         canonical.ID.String(),
		Name:       canonical.DisplayName,
		IsApproved: canonical.IsApproved,
		Aliases:    aliases,
	}
}

// ============================================================
// Tag Merge Operations Implementation
// ============================================================

// MergeTags manually merges source tag into target tag
// Business logic:
// 1. Validate source != target
// 2. Validate both tags exist
// 3. Move all aliases from source to target
// 4. Update video relationships
// 5. Delete source canonical tag
func (s *tagServiceV2) MergeTags(ctx context.Context, req dto.MergeTagsRequest) (*dto.MergeTagsResponse, error) {
	fmt.Printf("\n[MERGE_TAGS] ========================================\n")
	fmt.Printf("[MERGE_TAGS] Source: %s, Target: %s\n", req.SourceID, req.TargetID)

	// Parse UUIDs
	sourceUUID, err := uuid.Parse(req.SourceID)
	if err != nil {
		return nil, fmt.Errorf("invalid source tag ID: %w", err)
	}

	targetUUID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("invalid target tag ID: %w", err)
	}

	// Validate source != target
	if sourceUUID == targetUUID {
		return nil, fmt.Errorf("source and target tags must be different")
	}

	// Validate both tags exist
	sourceTag, err := s.tagRepo.GetCanonicalByID(ctx, sourceUUID)
	if err != nil {
		return nil, fmt.Errorf("source tag not found: %w", err)
	}

	targetTag, err := s.tagRepo.GetCanonicalByID(ctx, targetUUID)
	if err != nil {
		return nil, fmt.Errorf("target tag not found: %w", err)
	}

	fmt.Printf("[MERGE_TAGS] Merging '%s' → '%s'\n", sourceTag.DisplayName, targetTag.DisplayName)

	// Perform merge operation (atomic transaction)
	mergedAliasCount, err := s.tagRepo.MergeTags(ctx, sourceUUID, targetUUID)
	if err != nil {
		fmt.Printf("[MERGE_TAGS] ✗ Failed: %v\n", err)
		return nil, fmt.Errorf("failed to merge tags: %w", err)
	}

	fmt.Printf("[MERGE_TAGS] ✓ Success: Merged %d aliases\n", mergedAliasCount)
	fmt.Printf("[MERGE_TAGS] ========================================\n\n")

	// Build response
	return &dto.MergeTagsResponse{
		TargetTag: dto.TagResponse{
			ID:         targetTag.ID.String(),
			Name:       targetTag.DisplayName,
			IsApproved: targetTag.IsApproved,
		},
		MergedAliasCount: mergedAliasCount,
		SourceTagDeleted: true,
	}, nil
}

// UpdateTagApproval updates the approval status of a canonical tag
func (s *tagServiceV2) UpdateTagApproval(ctx context.Context, tagID string, isApproved bool) (*dto.TagResponse, error) {
	tagUUID, err := uuid.Parse(tagID)
	if err != nil {
		return nil, fmt.Errorf("invalid tag ID: %w", err)
	}

	// Get the tag first to ensure it exists
	tag, err := s.tagRepo.GetCanonicalByID(ctx, tagUUID)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// Update approval status
	tag.IsApproved = isApproved
	if err := s.tagRepo.UpdateCanonicalTag(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to update tag approval: %w", err)
	}

	return &dto.TagResponse{
		ID:         tag.ID.String(),
		Name:       tag.DisplayName,
		IsApproved: tag.IsApproved,
	}, nil
}
