package domain

import (
	"api/internal/dto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService interface {
	Login(req dto.LoginRequest, userAgent, clientIP string) (*dto.AuthResponse, error)
	Signup(req dto.SignupRequest, userAgent, clientIP string) (*dto.AuthResponse, error)
	VerifyToken(tokenString string) (*jwt.MapClaims, error)
	RefreshToken(refreshToken string) (*dto.AuthResponse, error)
	Logout(sessionID uuid.UUID) error
	LogoutAll(userID uuid.UUID) error

	// Google OAuth
	GetGoogleAuthURL(state string) string
	GenerateStateToken() string
	HandleGoogleCallback(code string, userAgent, clientIP string) (*dto.AuthResponse, error)

	// Session management
	CreateSession(userID uuid.UUID, userAgent, clientIP string) (*Session, string, error)
	ValidateSession(sessionID uuid.UUID) (*Session, error)
	GetSessionByRefreshToken(token string) (*Session, error)

	// Profile management
	UpdateMe(userID uuid.UUID, req dto.UpdateMeRequest) (*dto.UserResponse, error)
}
