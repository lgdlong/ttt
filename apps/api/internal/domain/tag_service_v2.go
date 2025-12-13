package domain

import (
	"api/internal/dto"
	"context"
)

// TagServiceV2 provides canonical-alias architecture API for tag management
// This service implements the new v2 API that replaces the legacy tag system
type TagServiceV2 interface {
	// ResolveTag implements 4-layer resolution to get or create canonical tag
	// Layer 1: Exact match (normalized text lookup)
	// Layer 2: Embedding generation
	// Layer 3: Semantic search (vector similarity)
	// Layer 4: Decision making (auto-merge or create new)
	// Returns: (CanonicalTag, matchedAlias, isNewTag, error)
	ResolveTag(ctx context.Context, userInput string) (*CanonicalTag, string, bool, error)

	// ============================================================
	// Canonical Tag CRUD
	// ============================================================

	GetCanonicalTagByID(ctx context.Context, id string) (*dto.TagResponse, error)
	ListCanonicalTags(ctx context.Context, req dto.TagListRequest) (*dto.TagListResponse, error)
	SearchCanonicalTags(ctx context.Context, query string, limit int) ([]dto.TagResponse, error)

	// ============================================================
	// Video-Canonical Tag Management
	// ============================================================

	// AddCanonicalTagToVideo adds a canonical tag to video (with auto-resolution)
	AddCanonicalTagToVideo(ctx context.Context, videoID string, req dto.AddVideoTagRequest) (*dto.TagResponse, error)
	RemoveCanonicalTagFromVideo(ctx context.Context, videoID, canonicalTagID string) error
	GetVideoCanonicalTags(ctx context.Context, videoID string) ([]dto.TagResponse, error)

	// ============================================================
	// Tag Merge Operations
	// ============================================================

	// MergeTags manually merges source tag into target tag
	// Source tag becomes an alias pointing to target canonical tag
	// Returns merged tag info and error
	MergeTags(ctx context.Context, req dto.MergeTagsRequest) (*dto.MergeTagsResponse, error)

	// ============================================================
	// Tag Approval Operations
	// ============================================================

	// UpdateTagApproval updates the approval status of a canonical tag
	UpdateTagApproval(ctx context.Context, tagID string, isApproved bool) (*dto.TagResponse, error)
}
