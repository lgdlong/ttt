package repository

import (
	"api/internal/domain"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagRepository interface {
	// Tag CRUD
	Create(tag *domain.Tag) error
	GetByID(id uuid.UUID) (*domain.Tag, error)
	GetByName(name string) (*domain.Tag, error)
	Update(tag *domain.Tag) error
	Delete(id uuid.UUID) error
	List(page, limit int) ([]domain.Tag, int64, error)
	Search(query string, limit int) ([]domain.Tag, error)

	// Video-Tag relationship
	AddTagToVideo(videoID, tagID uuid.UUID) error
	RemoveTagFromVideo(videoID, tagID uuid.UUID) error
	GetTagsByVideoID(videoID uuid.UUID) ([]domain.Tag, error)
	GetOrCreateByName(name string) (*domain.Tag, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// Create creates a new tag
func (r *tagRepository) Create(tag *domain.Tag) error {
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

// Search searches tags by name
func (r *tagRepository) Search(query string, limit int) ([]domain.Tag, error) {
	var tags []domain.Tag
	if err := r.db.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%").
		Order("name ASC").
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, err
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

// GetOrCreateByName gets a tag by name or creates it if not exists
func (r *tagRepository) GetOrCreateByName(name string) (*domain.Tag, error) {
	tag, err := r.GetByName(name)
	if err == nil {
		return tag, nil
	}

	// Create new tag
	newTag := &domain.Tag{
		Name: name,
	}
	if err := r.Create(newTag); err != nil {
		return nil, err
	}
	return newTag, nil
}
