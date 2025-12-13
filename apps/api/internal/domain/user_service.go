package domain

import (
	"api/internal/dto"
)

type UserService interface {
	// CRUD operations
	CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(id string) (*dto.UserResponse, error)
	UpdateUser(id string, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(id string) error
	ListUsers(req dto.ListUserRequest) (*dto.UserListResponse, error)
}
