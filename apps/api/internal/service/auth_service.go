package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/repository"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserInfo represents the user info from Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type AuthService interface {
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
	Signup(req dto.SignupRequest) (*dto.AuthResponse, error)
	VerifyToken(tokenString string) (*jwt.MapClaims, error)
	RefreshToken(refreshToken string) (*dto.AuthResponse, error)
	Logout(sessionID uuid.UUID) error
	LogoutAll(userID uuid.UUID) error

	// Google OAuth
	GetGoogleAuthURL(state string) string
	GenerateStateToken() string
	HandleGoogleCallback(code string, userAgent, clientIP string) (*dto.AuthResponse, error)

	// Session management
	CreateSession(userID uuid.UUID, userAgent, clientIP string) (*domain.Session, string, error)
	ValidateSession(sessionID uuid.UUID) (*domain.Session, error)
}

type authService struct {
	userRepo          repository.UserRepository
	socialAccountRepo repository.SocialAccountRepository
	sessionRepo       repository.SessionRepository
	jwtSecret         string
	googleOAuthConfig *oauth2.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	socialAccountRepo repository.SocialAccountRepository,
	sessionRepo repository.SessionRepository,
) AuthService {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production" // Fallback for development
	}

	// Google OAuth config
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleCallbackURL := os.Getenv("GOOGLE_CALLBACK_URL")
	if googleCallbackURL == "" {
		googleCallbackURL = "http://localhost:8080/api/auth/google/callback"
	}

	googleOAuthConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleCallbackURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &authService{
		userRepo:          userRepo,
		socialAccountRepo: socialAccountRepo,
		sessionRepo:       sessionRepo,
		jwtSecret:         jwtSecret,
		googleOAuthConfig: googleOAuthConfig,
	}
}

// Login authenticates user with username and password
func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by username
	user, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		User:  *s.toUserResponse(user),
		Token: token,
	}, nil
}

// Signup creates a new user account
func (s *authService) Signup(req dto.SignupRequest) (*dto.AuthResponse, error) {
	// Check if username already exists
	if _, err := s.userRepo.GetUserByUsername(req.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	if _, err := s.userRepo.GetUserByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user with default role and active status
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         string(domain.UserRoleUser), // Default: user
		IsActive:     true,                        // Default: active
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		User:  *s.toUserResponse(user),
		Token: token,
	}, nil
}

// VerifyToken validates JWT token and returns claims
func (s *authService) VerifyToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// RefreshToken generates a new access token using refresh token
func (s *authService) RefreshToken(refreshToken string) (*dto.AuthResponse, error) {
	// Find session by refresh token
	session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if session is blocked
	if session.IsBlocked {
		return nil, errors.New("session is blocked")
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("session expired")
	}

	// Get user
	user, err := s.userRepo.GetUserByID(session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new access token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		User:  *s.toUserResponse(user),
		Token: token,
	}, nil
}

// Logout invalidates a specific session
func (s *authService) Logout(sessionID uuid.UUID) error {
	return s.sessionRepo.BlockSession(sessionID)
}

// LogoutAll invalidates all sessions for a user
func (s *authService) LogoutAll(userID uuid.UUID) error {
	return s.sessionRepo.BlockAllUserSessions(userID)
}

// GetGoogleAuthURL returns the Google OAuth URL
func (s *authService) GetGoogleAuthURL(state string) string {
	return s.googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GenerateStateToken generates a random state token for CSRF protection
func (s *authService) GenerateStateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// HandleGoogleCallback processes the Google OAuth callback
func (s *authService) HandleGoogleCallback(code string, userAgent, clientIP string) (*dto.AuthResponse, error) {
	// Exchange code for token
	token, err := s.googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	googleUser, err := s.getGoogleUserInfo(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if social account exists
	socialAccount, err := s.socialAccountRepo.GetByProviderAndSocialID("google", googleUser.ID)
	if err == nil {
		// Social account exists, get user
		user, err := s.userRepo.GetUserByID(socialAccount.UserID)
		if err != nil {
			return nil, errors.New("user not found")
		}

		// Check if user is active
		if !user.IsActive {
			return nil, errors.New("account is deactivated")
		}

		// Generate token
		jwtToken, err := s.generateToken(user)
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}

		return &dto.AuthResponse{
			User:  *s.toUserResponse(user),
			Token: jwtToken,
		}, nil
	}

	// Social account doesn't exist, check if user with email exists
	existingUser, err := s.userRepo.GetUserByEmail(googleUser.Email)
	if err == nil {
		// User exists, link social account
		socialAccount := &domain.SocialAccount{
			UserID:   existingUser.ID,
			Provider: "google",
			SocialID: googleUser.ID,
			Email:    googleUser.Email,
		}

		if err := s.socialAccountRepo.Create(socialAccount); err != nil {
			return nil, fmt.Errorf("failed to link social account: %w", err)
		}

		// Check if user is active
		if !existingUser.IsActive {
			return nil, errors.New("account is deactivated")
		}

		// Generate token
		jwtToken, err := s.generateToken(existingUser)
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}

		return &dto.AuthResponse{
			User:  *s.toUserResponse(existingUser),
			Token: jwtToken,
		}, nil
	}

	// Create new user
	username := s.generateUsernameFromEmail(googleUser.Email)
	newUser := &domain.User{
		Username:     username,
		Email:        googleUser.Email,
		PasswordHash: "", // No password for OAuth users
		Role:         string(domain.UserRoleUser),
		IsActive:     true,
	}

	if err := s.userRepo.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create social account link
	socialAccount = &domain.SocialAccount{
		UserID:   newUser.ID,
		Provider: "google",
		SocialID: googleUser.ID,
		Email:    googleUser.Email,
	}

	if err := s.socialAccountRepo.Create(socialAccount); err != nil {
		return nil, fmt.Errorf("failed to create social account: %w", err)
	}

	// Generate token
	jwtToken, err := s.generateToken(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		User:  *s.toUserResponse(newUser),
		Token: jwtToken,
	}, nil
}

// CreateSession creates a new session for the user
func (s *authService) CreateSession(userID uuid.UUID, userAgent, clientIP string) (*domain.Session, string, error) {
	// Generate refresh token
	refreshToken := s.generateRefreshToken()

	session := &domain.Session{
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 30), // 30 days
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return session, refreshToken, nil
}

// ValidateSession validates a session by ID
func (s *authService) ValidateSession(sessionID uuid.UUID) (*domain.Session, error) {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	if session.IsBlocked {
		return nil, errors.New("session is blocked")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("session expired")
	}

	return session, nil
}

// getGoogleUserInfo fetches user info from Google API
func (s *authService) getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// generateToken creates a JWT token for the user
func (s *authService) generateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 hours expiration (shorter for access token)
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken generates a secure random refresh token
func (s *authService) generateRefreshToken() string {
	b := make([]byte, 64)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// generateUsernameFromEmail creates a username from email
func (s *authService) generateUsernameFromEmail(email string) string {
	// Take part before @
	parts := strings.Split(email, "@")
	username := parts[0]

	// Check if username exists, if so, append random suffix
	if _, err := s.userRepo.GetUserByUsername(username); err == nil {
		b := make([]byte, 4)
		rand.Read(b)
		username = username + "_" + base64.RawURLEncoding.EncodeToString(b)[:6]
	}

	return username
}

// Helper: Convert domain.User to dto.UserResponse
func (s *authService) toUserResponse(user *domain.User) *dto.UserResponse {
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
