package domain

import (
	"github.com/google/uuid"
)

type SocialAccountRepository interface {
	Create(account *SocialAccount) error
	GetByProviderAndSocialID(provider, socialID string) (*SocialAccount, error)
	GetByUserID(userID uuid.UUID) ([]SocialAccount, error)
	Delete(id uuid.UUID) error
}
