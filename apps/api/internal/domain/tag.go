package domain

import (
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Tag struct {
	ID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name string    `gorm:"type:varchar(100);uniqueIndex;not null"`

	// vector(1536) tương thích với OpenAI text-embedding-3-small
	Embedding pgvector.Vector `gorm:"type:vector(1536)"`
}

func (Tag) TableName() string {
	return "tags"
}
