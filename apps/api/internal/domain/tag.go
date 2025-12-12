package domain

import (
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

// GenerateSlug creates URL-friendly slug from display name
// Example: "Money & Finance" → "money-finance"
func GenerateSlug(displayName string) string {
	return slug.Make(displayName)
}

// NewCanonicalTag creates a new canonical tag with auto-generated slug
func NewCanonicalTag(displayName string) *CanonicalTag {
	return &CanonicalTag{
		DisplayName: displayName,
		Slug:        GenerateSlug(displayName),
	}
}

// NewTagAlias creates a new tag alias with normalized text
func NewTagAlias(rawText string, canonicalTagID uuid.UUID, embedding pgvector.Vector, similarityScore float64) *TagAlias {
	return &TagAlias{
		CanonicalTagID:  canonicalTagID,
		RawText:         rawText,
		NormalizedText:  NormalizeText(rawText),
		Embedding:       embedding,
		SimilarityScore: similarityScore,
		Language:        "unk", // TODO: Detect language using external library
	}
}
