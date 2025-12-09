package repository

import (
	"api/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SocialAccountRepository interface {
	Create(account *domain.SocialAccount) error
	GetByProviderAndSocialID(provider, socialID string) (*domain.SocialAccount, error)
	GetByUserID(userID uuid.UUID) ([]domain.SocialAccount, error)
	Delete(id uuid.UUID) error
}

type socialAccountRepository struct {
	db *gorm.DB
}

func NewSocialAccountRepository(db *gorm.DB) SocialAccountRepository {
	return &socialAccountRepository{db: db}
}

// Create creates a new social account link
func (r *socialAccountRepository) Create(account *domain.SocialAccount) error {
	return r.db.Create(account).Error
}

// GetByProviderAndSocialID finds a social account by provider and social ID
func (r *socialAccountRepository) GetByProviderAndSocialID(provider, socialID string) (*domain.SocialAccount, error) {
	var account domain.SocialAccount
	if err := r.db.Preload("User").
		Where("provider = ? AND social_id = ?", provider, socialID).
		First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByUserID returns all social accounts for a user
func (r *socialAccountRepository) GetByUserID(userID uuid.UUID) ([]domain.SocialAccount, error) {
	var accounts []domain.SocialAccount
	if err := r.db.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// Delete removes a social account link
func (r *socialAccountRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.SocialAccount{}, "id = ?", id).Error
}
