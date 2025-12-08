package repository

import (
	"api/internal/domain"
	"api/internal/dto"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	// CRUD operations
	CreateUser(user *domain.User) error
	GetUserByID(id uuid.UUID) (*domain.User, error)
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	UpdateUser(id uuid.UUID, updates map[string]interface{}) error
	DeleteUser(id uuid.UUID) error // Soft delete (GORM handles via DeletedAt)
	ListUsers(req dto.ListUserRequest) ([]domain.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser creates a new user
func (r *userRepository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

// GetUserByID retrieves a user by ID (GORM auto-excludes soft deleted)
func (r *userRepository) GetUserByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *userRepository) GetUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *userRepository) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user fields
func (r *userRepository) UpdateUser(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteUser performs soft delete (GORM handles via DeletedAt)
func (r *userRepository) DeleteUser(id uuid.UUID) error {
	return r.db.Delete(&domain.User{}, "id = ?", id).Error
}

// ListUsers retrieves paginated list of users with optional filters
func (r *userRepository) ListUsers(req dto.ListUserRequest) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.Model(&domain.User{})

	// Apply role filter if provided
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	// Apply is_active filter if provided
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (req.Page - 1) * req.Limit
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}
