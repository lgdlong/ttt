package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/repository"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	// CRUD operations
	CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(id string) (*dto.UserResponse, error)
	UpdateUser(id string, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(id string) error
	ListUsers(req dto.ListUserRequest) (*dto.UserListResponse, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// CreateUser creates a new user with hashed password
func (s *userService) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Check if username already exists
	if _, err := s.repo.GetUserByUsername(req.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	if _, err := s.repo.GetUserByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = string(domain.UserRoleUser)
	}

	// Create user
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		IsActive:     true, // Default: active
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(id string) (*dto.UserResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.toUserResponse(user), nil
}

// UpdateUser updates user information
func (s *userService) UpdateUser(id string, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Username != nil {
		// Check if new username already exists
		if existingUser, err := s.repo.GetUserByUsername(*req.Username); err == nil && existingUser.ID != userID {
			return nil, errors.New("username already exists")
		}
		updates["username"] = *req.Username
	}

	if req.Email != nil {
		// Check if new email already exists
		if existingUser, err := s.repo.GetUserByEmail(*req.Email); err == nil && existingUser.ID != userID {
			return nil, errors.New("email already exists")
		}
		updates["email"] = *req.Email
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		updates["password_hash"] = string(hashedPassword)
	}

	if req.Role != nil {
		updates["role"] = *req.Role
	}

	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	updates["updated_at"] = time.Now()

	if err := s.repo.UpdateUser(userID, updates); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Fetch updated user
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// DeleteUser performs soft delete
func (s *userService) DeleteUser(id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	if err := s.repo.DeleteUser(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves paginated list of users
func (s *userService) ListUsers(req dto.ListUserRequest) (*dto.UserListResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	users, total, err := s.repo.ListUsers(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to DTOs
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.toUserResponse(&user)
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &dto.UserListResponse{
		Data: userResponses,
		Pagination: dto.PaginationMetadata{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

// Helper: Convert domain.User to dto.UserResponse
func (s *userService) toUserResponse(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
