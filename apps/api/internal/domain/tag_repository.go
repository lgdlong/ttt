package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type TagRepository interface {
	// ============================================================
	// Canonical-Alias Architecture (New)
	// ============================================================

	// GetCanonicalByAlias finds canonical tag by exact normalized text match (Layer 1)
	// Example: "tiền" → CanonicalTag{ID: uuid, DisplayName: "Money"}
	// Returns nil if not found (not an error, proceed to Layer 2)
	GetCanonicalByAlias(ctx context.Context, normalizedText string) (*CanonicalTag, error)

	// GetClosestCanonical finds most similar canonical tag using vector search (Layer 3)
	// Returns (CanonicalTag, similarity_score, error)
	// If no match above threshold, returns (nil, 0, nil)
	GetClosestCanonical(ctx context.Context, embedding pgvector.Vector, threshold float64) (*CanonicalTag, float64, error)

	// CreateCanonicalTag creates a new canonical tag with its initial alias (atomic transaction)
	// Used when: No similar tag found (Layer 4 - Scenario B)
	CreateCanonicalTag(ctx context.Context, canonical *CanonicalTag, initialAlias *TagAlias) error

	// CreateAlias adds a new alias to existing canonical tag
	// Used when: Similar tag found (Layer 4 - Scenario A)
	CreateAlias(ctx context.Context, alias *TagAlias) error

	// GetCanonicalByID retrieves canonical tag by ID
	GetCanonicalByID(ctx context.Context, id uuid.UUID) (*CanonicalTag, error)

	// UpdateCanonicalTag updates a canonical tag
	UpdateCanonicalTag(ctx context.Context, canonical *CanonicalTag) error

	// ListCanonicalTags returns paginated list of canonical tags
	ListCanonicalTags(ctx context.Context, page, limit int) ([]CanonicalTag, int64, error)

	// SearchCanonicalTags searches canonical tags (hybrid: SQL LIKE + Vector)
	SearchCanonicalTags(ctx context.Context, query string, limit int) ([]CanonicalTag, error)

	// ============================================================
	// Video-Canonical Tag Relationship
	// ============================================================

	// AddCanonicalTagToVideo links a canonical tag to a video
	AddCanonicalTagToVideo(ctx context.Context, videoID, canonicalTagID uuid.UUID) error

	// RemoveCanonicalTagFromVideo unlinks a canonical tag from a video
	RemoveCanonicalTagFromVideo(ctx context.Context, videoID, canonicalTagID uuid.UUID) error

	// GetCanonicalTagsByVideoID returns all canonical tags for a video
	GetCanonicalTagsByVideoID(ctx context.Context, videoID uuid.UUID) ([]CanonicalTag, error)

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
