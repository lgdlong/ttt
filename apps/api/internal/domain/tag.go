package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/pgvector/pgvector-go"
)

// CanonicalTag đại diện cho một chủ đề duy nhất (concept)
// Ví dụ: ID=1, Slug="money-finance", DisplayName="Money"
type CanonicalTag struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Slug        string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	DisplayName string    `gorm:"type:varchar(100);not null"`

	// Approval status - tags need to be approved by moderator before being visible
	IsApproved bool `gorm:"default:false"`

	// Has Many Aliases (one-to-many relationship)
	Aliases []TagAlias `gorm:"foreignKey:CanonicalTagID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TagAlias là các từ khóa khác nhau trỏ về cùng một CanonicalTag
// Ví dụ: RawText="Tiền", NormalizedText="tiền", CanonicalTagID=1 (Money)
type TagAlias struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CanonicalTagID uuid.UUID `gorm:"type:uuid;not null;index"`

	// User input variants
	RawText        string `gorm:"type:varchar(100);not null"`
	NormalizedText string `gorm:"type:varchar(100);not null;uniqueIndex"` // LOWER(TRIM(raw_text))
	Language       string `gorm:"type:varchar(10);default:'unk'"`

	// Vector Embedding for semantic search (text-embedding-3-small: 1536 dims)
	Embedding pgvector.Vector `gorm:"type:vector(1536)"`

	// Metadata for admin review
	IsReviewed      bool    `gorm:"default:false"` // FALSE = AI auto-mapped
	SimilarityScore float64 `gorm:"type:float;default:1.0"`

	CreatedAt time.Time
}

func (CanonicalTag) TableName() string { return "canonical_tags" }
func (TagAlias) TableName() string     { return "tag_aliases" }

// ============================================================

// NormalizeText converts user input to normalized form for exact matching
// Example: "  Tiền  " → "tiền"
func NormalizeText(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

// GenerateSlug creates URL-friendly slug from display name with validation and truncation
// Example: "Money & Finance" → "money-finance"
// Returns empty string if input is invalid (caller should handle this)
func GenerateSlug(displayName string) string {
	const maxSlugLength = 95 // Reserve 5 chars for potential collision suffix (e.g., "-1234")

	// Validate input
	trimmed := strings.TrimSpace(displayName)
	if trimmed == "" {
		return "" // Return empty string for invalid input
	}

	// Generate slug using gosimple/slug library
	generated := slug.Make(trimmed)

	// Guard against empty slug generation (e.g., non-ASCII with no fallback)
	if generated == "" {
		// Fallback: use normalized lowercase version
		generated = strings.ToLower(strings.ReplaceAll(trimmed, " ", "-"))
	}

	// Truncate to fit DB constraint (varchar(100)) with room for collision suffix
	if len(generated) > maxSlugLength {
		generated = generated[:maxSlugLength]
		// Trim trailing dashes after truncation
		generated = strings.TrimRight(generated, "-")
	}

	// TODO: Collision Handling
	// If slug already exists in DB, append numeric suffix: "tag-name-1", "tag-name-2", etc.
	// This should be handled at the repository/service layer by:
	// 1. Checking if slug exists: SELECT COUNT(*) FROM canonical_tags WHERE slug = ?
	// 2. If exists, append "-{counter}" and retry until unique
	// 3. Wrap in transaction to prevent race conditions
	// Example implementation:
	//   baseSlug := GenerateSlug(displayName)
	//   finalSlug := baseSlug
	//   counter := 1
	//   for slugExists(finalSlug) {
	//       finalSlug = fmt.Sprintf("%s-%d", baseSlug, counter)
	//       counter++
	//   }

	return generated
}

// NewCanonicalTag creates a new canonical tag with auto-generated slug
// Returns error if displayName is empty or slug generation fails
func NewCanonicalTag(displayName string) (*CanonicalTag, error) {
	if strings.TrimSpace(displayName) == "" {
		return nil, fmt.Errorf("displayName cannot be empty")
	}

	slug := GenerateSlug(displayName)
	if slug == "" {
		return nil, fmt.Errorf("failed to generate valid slug from displayName: %s", displayName)
	}

	return &CanonicalTag{
		DisplayName: displayName,
		Slug:        slug,
	}, nil
}

// NewTagAlias creates a new tag alias with normalized text
// Returns error if validation fails (nil canonicalTagID or empty rawText)
func NewTagAlias(rawText string, canonicalTagID uuid.UUID, embedding pgvector.Vector, similarityScore float64) (*TagAlias, error) {
	// Validate inputs
	if canonicalTagID == uuid.Nil {
		return nil, fmt.Errorf("canonicalTagID cannot be nil")
	}
	if strings.TrimSpace(rawText) == "" {
		return nil, fmt.Errorf("rawText cannot be empty")
	}

	return &TagAlias{
		CanonicalTagID:  canonicalTagID,
		RawText:         rawText,
		NormalizedText:  NormalizeText(rawText),
		Embedding:       embedding,
		SimilarityScore: similarityScore,
		Language:        "unk", // TODO: Detect language using external library
	}, nil
}

// NewInitialTagAlias creates a tag alias for initial canonical creation
// Used when creating both canonical and alias together - the repository will set CanonicalTagID
// Returns error if rawText validation fails
func NewInitialTagAlias(rawText string, embedding pgvector.Vector, similarityScore float64) (*TagAlias, error) {
	if strings.TrimSpace(rawText) == "" {
		return nil, fmt.Errorf("rawText cannot be empty")
	}

	return &TagAlias{
		CanonicalTagID:  uuid.Nil, // Will be set by repository after canonical creation
		RawText:         rawText,
		NormalizedText:  NormalizeText(rawText),
		Embedding:       embedding,
		SimilarityScore: similarityScore,
		Language:        "unk",
	}, nil
}
