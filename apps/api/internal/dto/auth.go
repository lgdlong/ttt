package dto

import "time"

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

// UpdateMeRequest - Request for a user to update their own profile info
type UpdateMeRequest struct {
	FullName *string `json:"full_name" binding:"omitempty,max=100"`
	Email    *string `json:"email" binding:"omitempty,email,max=100"`
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