package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"context"
	"fmt"
	"log/slog"
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
	slog.Info("Tag resolution started", "input", userInput)

	// ============================================================
	// Layer 1: Exact Match (Normalized Text Lookup)
	// Cost: ~10ms, $0
	// ============================================================
	normalizedInput := domain.NormalizeText(userInput)
	slog.Debug("Layer 1: Checking exact match", "normalized_input", normalizedInput)

	canonical, err := s.tagRepo.GetCanonicalByAlias(ctx, normalizedInput)
	if err != nil {
		return nil, "", false, fmt.Errorf("Layer 1 failed: %w", err)
	}

	if canonical != nil {
		slog.Info("Layer 1 hit: Found exact match",
			"action", "return_existing_canonical",
			"input", userInput,
			"canonical_name", canonical.DisplayName,
			"canonical_id", canonical.ID.String(),
		)
		return canonical, normalizedInput, false, nil
	}

	slog.Debug("Layer 1 miss: No exact match found", "normalized_input", normalizedInput)

	// ============================================================
	// Layer 1.5: Translation Layer (Cross-Lingual Support)
	// Cost: ~300ms, ~$0.00015 (GPT-4o-mini API call)
	// Purpose: Normalize non-English queries to English before vector search
	// Example: "Tiền" -> "Money", "Học máy" -> "Machine Learning"
	// ============================================================
	slog.Debug("Layer 1.5: Attempting translation to English", "input", userInput)

	// Attempt translation to English
	englishTerm, err := s.tagRepo.TranslateText(ctx, userInput)

	// Process translation result only if successful and different from original
	if err == nil && englishTerm != "" && domain.NormalizeText(englishTerm) != normalizedInput {
		slog.Debug("Layer 1.5: Translation successful",
			"original", userInput,
			"translated", englishTerm,
		)

		// Check if translated English term has existing canonical tag
		normalizedEng := domain.NormalizeText(englishTerm)
		canonicalEng, err := s.tagRepo.GetCanonicalByAlias(ctx, normalizedEng)

		if err == nil && canonicalEng != nil {
			slog.Info("Layer 1.5 hit: Found canonical via translation",
				"action", "return_via_translation",
				"original_input", userInput,
				"translated_term", englishTerm,
				"canonical_name", canonicalEng.DisplayName,
				"canonical_id", canonicalEng.ID.String(),
			)

			// Auto-create alias for original input to avoid future translation costs
			// Use original input's embedding (not translated term) for future semantic search
			embeddingSlice, embErr := s.tagRepo.GetEmbeddingForText(ctx, userInput)
			if embErr == nil {
				embedding := pgvector.NewVector(embeddingSlice)
				newAlias, aliasErr := domain.NewTagAlias(userInput, canonicalEng.ID, embedding, 1.0)
				if aliasErr != nil {
					slog.Warn("Failed to create translation alias (validation)",
						"original_input", userInput,
						"error", aliasErr.Error(),
					)
				} else {
					// Save alias to optimize future lookups (next time will hit Layer 1)
					if createErr := s.tagRepo.CreateAlias(ctx, newAlias); createErr != nil {
						// Log warning but don't fail request - still return found canonical
						slog.Warn("Failed to save translation alias (DB)",
							"original_input", userInput,
							"error", createErr.Error(),
						)
					} else {
						slog.Debug("Translation alias created",
							"original", userInput,
							"canonical", canonicalEng.DisplayName,
						)
					}
				}
			}

			return canonicalEng, userInput, false, nil
		}

		slog.Debug("Layer 1.5 miss: Translated term not found",
			"original", userInput,
			"translated", englishTerm,
		)
	} else if err != nil {
		slog.Warn("Layer 1.5: Translation failed (continuing to Layer 2)",
			"input", userInput,
			"error", err.Error(),
		)
	} else {
		slog.Debug("Layer 1.5 skipped: Input already in English or same translation")
	}

	// ============================================================
	// Layer 2: Embedding Generation
	// Cost: ~500ms, ~$0.0001 (OpenAI API call)
	// ============================================================
	slog.Debug("Layer 2: Generating embedding", "input", userInput)

	embeddingSlice, err := s.tagRepo.GetEmbeddingForText(ctx, userInput)
	if err != nil {
		// OpenAI unavailable → Create new canonical without semantic check
		slog.Warn("Layer 2 failed: OpenAI unavailable, creating canonical without embedding",
			"input", userInput,
			"error", err.Error(),
		)

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

		slog.Info("Layer 2 fallback: Created new canonical (no embedding)",
			"action", "create_new_canonical_no_embedding",
			"input", userInput,
			"canonical_name", newCanonical.DisplayName,
			"canonical_id", newCanonical.ID.String(),
		)
		return newCanonical, userInput, true, nil
	}

	embedding := pgvector.NewVector(embeddingSlice)
	slog.Debug("Layer 2 success: Embedding generated",
		"input", userInput,
		"embedding_dims", len(embeddingSlice),
	)

	// ============================================================
	// Layer 3: Semantic Search (Vector Similarity)
	// Cost: ~50ms, $0 (uses cached embeddings in DB)
	// ============================================================
	slog.Debug("Layer 3: Performing semantic search",
		"input", userInput,
		"threshold", AUTO_MERGE_THRESHOLD,
	)

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
		slog.Info("Layer 3 hit: Found similar canonical",
			"action", "auto_merge_to_existing",
			"input", userInput,
			"matched_canonical", closestCanonical.DisplayName,
			"matched_id", closestCanonical.ID.String(),
			"similarity_score", similarityScore,
		)

		// Create new alias pointing to existing canonical
		newAlias, aliasErr := domain.NewTagAlias(userInput, closestCanonical.ID, embedding, similarityScore)
		if aliasErr != nil {
			return nil, "", false, fmt.Errorf("failed to create alias (validation failed): %w", aliasErr)
		}
		if err := s.tagRepo.CreateAlias(ctx, newAlias); err != nil {
			return nil, "", false, fmt.Errorf("failed to create alias: %w", err)
		}

		slog.Debug("Auto-merge alias created",
			"original_input", userInput,
			"matched_canonical", closestCanonical.DisplayName,
		)
		return closestCanonical, userInput, false, nil
	}

	// Scenario B: No Match (Score < Threshold)
	// Action: Create new canonical + initial alias
	slog.Info("Layer 3 miss: No similar canonical found, creating new",
		"action", "create_new_canonical",
		"input", userInput,
		"threshold", AUTO_MERGE_THRESHOLD,
	)

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

	slog.Info("Tag resolution complete: New canonical created",
		"action", "create_new_canonical_with_embedding",
		"input", userInput,
		"canonical_name", newCanonical.DisplayName,
		"canonical_id", newCanonical.ID.String(),
		"embedding_dims", len(embeddingSlice),
	)
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
	slog.Info("Tag merge operation started",
		"action", "merge_tags",
		"source_id", req.SourceID,
		"target_id", req.TargetID,
	)

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

	slog.Debug("Validated merge request",
		"source_name", sourceTag.DisplayName,
		"target_name", targetTag.DisplayName,
	)

	// Perform merge operation (atomic transaction)
	mergedAliasCount, err := s.tagRepo.MergeTags(ctx, sourceUUID, targetUUID)
	if err != nil {
		slog.Error("Tag merge operation failed",
			"action", "merge_tags",
			"source_id", req.SourceID,
			"target_id", req.TargetID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to merge tags: %w", err)
	}

	slog.Info("Tag merge operation completed",
		"action", "merge_tags",
		"source_name", sourceTag.DisplayName,
		"target_name", targetTag.DisplayName,
		"merged_alias_count", mergedAliasCount,
		"source_id", req.SourceID,
		"target_id", req.TargetID,
	)

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
	slog.Debug("Tag approval update requested",
		"tag_id", tagID,
		"is_approved", isApproved,
	)

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
		slog.Error("Tag approval update failed",
			"tag_id", tagID,
			"tag_name", tag.DisplayName,
			"is_approved", isApproved,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to update tag approval: %w", err)
	}

	slog.Info("Tag approval updated",
		"action", "update_tag_approval",
		"tag_id", tagID,
		"tag_name", tag.DisplayName,
		"is_approved", isApproved,
	)

	return &dto.TagResponse{
		ID:         tag.ID.String(),
		Name:       tag.DisplayName,
		IsApproved: tag.IsApproved,
	}, nil
}
