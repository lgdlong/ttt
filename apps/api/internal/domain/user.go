package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
	UserRoleMod   UserRole = "mod"
)

type User struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	Username     string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Email        string `gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	FullName     string `gorm:"type:varchar(100);default:'';not null"`

	Role     string `gorm:"type:varchar(20);default:'user';not null"` // e.g., 'user', 'admin', 'mod'
	IsActive bool   `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft Delete (Xóa mềm)
}

func (User) TableName() string {
	return "users"
}
