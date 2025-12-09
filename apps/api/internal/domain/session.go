package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`

	// Token xoay vòng (Refresh Token)
	RefreshToken string `gorm:"not null;index" json:"-"`

	// Thông tin thiết bị (để user biết mình đang đăng nhập ở đâu)
	UserAgent string `json:"user_agent"` // Chrome on Windows...
	ClientIP  string `json:"client_ip"`

	IsBlocked bool      `gorm:"default:false" json:"is_blocked"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`

	CreatedAt time.Time `json:"created_at"`
}
