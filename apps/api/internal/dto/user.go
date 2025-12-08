package dto

import "time"

// ============ User CRUD DTOs ============

// CreateUserRequest - Request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"omitempty,max=100"`
	Role     string `json:"role" binding:"omitempty,oneof=user admin mod"`
}

// UpdateUserRequest - Request to update user information
type UpdateUserRequest struct {
	Username *string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    *string `json:"email" binding:"omitempty,email,max=100"`
	Password *string `json:"password" binding:"omitempty,min=6"`
	FullName *string `json:"full_name" binding:"omitempty,max=100"`
	Role     *string `json:"role" binding:"omitempty,oneof=user admin mod"`
	IsActive *bool   `json:"is_active" binding:"omitempty"`
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
}

// UserListResponse - Response with list of users and pagination
type UserListResponse struct {
	Data       []UserResponse     `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// ============ Authentication DTOs ============

// LoginRequest - Login with username and password
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SignupRequest - Signup with username, email and password
type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"omitempty,max=100"`
}

// AuthResponse - Response after successful login/signup
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token,omitempty"` // JWT token (omitted when using cookies)
}

// GoogleAuthURLResponse - Response with Google OAuth URL
type GoogleAuthURLResponse struct {
	URL string `json:"url"`
}

// GoogleAuthRequest - Request for Google OAuth login
type GoogleAuthRequest struct {
	IDToken string `json:"id_token" binding:"required"` // Google ID token from frontend
}

// GoogleUserInfo - User info from Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

// ============ Session DTOs ============

// SessionResponse - Session data for API responses
type SessionResponse struct {
	ID        string    `json:"id"`
	UserAgent string    `json:"user_agent"`
	ClientIP  string    `json:"client_ip"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsBlocked bool      `json:"is_blocked"`
}

// SocialAccountResponse - Social account data for API responses
type SocialAccountResponse struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
