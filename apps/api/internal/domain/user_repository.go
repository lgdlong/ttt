package domain

import (
	"api/internal/dto"

	"github.com/google/uuid"
)

type UserRepository interface {
	// CRUD operations
	CreateUser(user *User) error
	GetUserByID(id uuid.UUID) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(id uuid.UUID, updates map[string]interface{}) error
	DeleteUser(id uuid.UUID) error // Soft delete (GORM handles via DeletedAt)
	ListUsers(req dto.ListUserRequest) ([]User, int64, error)
}
