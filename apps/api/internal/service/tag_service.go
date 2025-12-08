package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/repository"
	"fmt"
	"math"

	"github.com/google/uuid"
)

type TagService interface {
	// Tag CRUD
	CreateTag(req dto.CreateTagRequest) (*dto.TagResponse, error)
	GetTagByID(id string) (*dto.TagResponse, error)
	UpdateTag(id string, req dto.UpdateTagRequest) (*dto.TagResponse, error)
	DeleteTag(id string) error
	ListTags(req dto.TagListRequest) (*dto.TagListResponse, error)
	SearchTags(query string, limit int) ([]dto.TagResponse, error)

	// Video-Tag management
	AddTagToVideo(videoID string, req dto.AddVideoTagRequest) (*dto.TagResponse, error)
	RemoveTagFromVideo(videoID, tagID string) error
	GetVideoTags(videoID string) ([]dto.TagResponse, error)
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

// CreateTag creates a new tag
func (s *tagService) CreateTag(req dto.CreateTagRequest) (*dto.TagResponse, error) {
	// Check if tag with same name exists
	if existing, _ := s.tagRepo.GetByName(req.Name); existing != nil {
		return nil, fmt.Errorf("tag with name '%s' already exists", req.Name)
	}

	tag := &domain.Tag{
		Name: req.Name,
	}

	if err := s.tagRepo.Create(tag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

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
func (s *tagService) ListTags(req dto.TagListRequest) (*dto.TagListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// If query provided, search instead of list
	if req.Query != "" {
		tags, err := s.tagRepo.Search(req.Query, req.Limit)
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

// SearchTags searches tags by name
func (s *tagService) SearchTags(query string, limit int) ([]dto.TagResponse, error) {
	if limit < 1 {
		limit = 20
	}

	tags, err := s.tagRepo.Search(query, limit)
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
func (s *tagService) AddTagToVideo(videoID string, req dto.AddVideoTagRequest) (*dto.TagResponse, error) {
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
		// Get or create tag by name
		tag, err = s.tagRepo.GetOrCreateByName(*req.TagName)
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
