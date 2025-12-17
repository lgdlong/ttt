package dto

import "time"

// ============ User CRUD DTOs ============

// CreateUserRequest - Request to create a new user
// Note: Role assignment is handled server-side. Use separate admin endpoints for role management.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"omitempty,max=100"`
}

// UpdateUserRequest - Request to update user information
// Note: Role and IsActive management require admin privileges. Use separate admin endpoints.
type UpdateUserRequest struct {
	Username *string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    *string `json:"email" binding:"omitempty,email,max=100"`
	Password *string `json:"password" binding:"omitempty,min=6"`
	FullName *string `json:"full_name" binding:"omitempty,max=100"`
}

// UserResponse - User data for API responses (without password)
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUserRequest - Request params for listing users
type ListUserRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1" default:"1"`
	Limit    int    `form:"limit" binding:"omitempty,min=1,max=100" default:"20"`
	Role     string `form:"role" binding:"omitempty,oneof=user admin mod"`
	IsActive *bool  `form:"is_active" binding:"omitempty"`
	Query    string `form:"q" binding:"omitempty"` // Search by username or email
}

// UserListResponse - Response with list of users and pagination
type UserListResponse struct {
	Data       []UserResponse     `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}
