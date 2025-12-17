package service

import (
	"api/internal/domain"
	"api/internal/dto"
	"api/internal/helper"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
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

const JWTDefaultExpiresIn = "24h"

type authService struct {
	userRepo              domain.UserRepository
	socialAccountRepo     domain.SocialAccountRepository
	sessionRepo           domain.SessionRepository
	jwtSecret             string
	jwtExpiresIn          time.Duration
	refreshTokenExpiresIn time.Duration // New field for refresh token expiration
	googleOAuthConfig     *oauth2.Config
}

func NewAuthService(
	userRepo domain.UserRepository,
	socialAccountRepo domain.SocialAccountRepository,
	sessionRepo domain.SessionRepository,
) domain.AuthService {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// The application should NEVER run with a default secret.
		// Fail fast if the secret is not configured.
		panic("FATAL: JWT_SECRET is not set")
	}

	// Parse JWT expiration duration from environment
	jwtExpiresInStr := os.Getenv("JWT_EXPIRES_IN")
	if jwtExpiresInStr == "" {
		jwtExpiresInStr = JWTDefaultExpiresIn // Default to 24 hours if not set
	}
	jwtExpiresIn, err := helper.ParseDurationWithDays(jwtExpiresInStr)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Invalid JWT_EXPIRES_IN format: %v", err))
	}

	// Parse Refresh Token expiration duration from environment
	refreshTokenExpiresInStr := os.Getenv("REFRESH_TOKEN_EXPIRES_IN")
	if refreshTokenExpiresInStr == "" {
		refreshTokenExpiresInStr = "720h" // Default to 30 days if not set
	}
	refreshTokenExpiresIn, err := helper.ParseDurationWithDays(refreshTokenExpiresInStr)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Invalid REFRESH_TOKEN_EXPIRES_IN format: %v", err))
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
		userRepo:              userRepo,
		socialAccountRepo:     socialAccountRepo,
		sessionRepo:           sessionRepo,
		jwtSecret:             jwtSecret,
		jwtExpiresIn:          jwtExpiresIn,
		refreshTokenExpiresIn: refreshTokenExpiresIn, // Assign the parsed duration
		googleOAuthConfig:     googleOAuthConfig,
	}
}

// Login authenticates user with username and password
func (s *authService) Login(req dto.LoginRequest, userAgent, clientIP string) (*dto.AuthResponse, error) {
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

	// Create session
	session, refreshToken, err := s.CreateSession(user.ID, userAgent, clientIP)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate JWT token with session ID
	token, err := s.generateToken(user, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		User:         *s.toUserResponse(user),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// Signup creates a new user account
func (s *authService) Signup(req dto.SignupRequest, userAgent, clientIP string) (*dto.AuthResponse, error) {
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
		FullName:     req.FullName,        // Optional full name
		Role:         domain.UserRoleUser, // Default: user
		IsActive:     true,                // Default: active
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create session
	session, refreshToken, err := s.CreateSession(user.ID, userAgent, clientIP)
	if err != nil {
		// Note: In a real-world scenario, you might want to handle the case
		// where user creation succeeds but session creation fails (e.g., by rolling back the user creation).
		// For this implementation, we'll return an error and the user will have to log in.
		return nil, fmt.Errorf("failed to create session after signup: %w", err)
	}

	// Generate JWT token with session ID
	token, err := s.generateToken(user, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		User:         *s.toUserResponse(user),
		Token:        token,
		RefreshToken: refreshToken,
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

	// Generate new access token, linked to the existing session
	token, err := s.generateToken(user, session.ID)
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

	// This is a helper function to reduce repetition
	createAuthResponse := func(user *domain.User) (*dto.AuthResponse, error) {
		if !user.IsActive {
			return nil, errors.New("account is deactivated")
		}

		session, refreshToken, err := s.CreateSession(user.ID, userAgent, clientIP)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}

		jwtToken, err := s.generateToken(user, session.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}

		return &dto.AuthResponse{
			User:         *s.toUserResponse(user),
			Token:        jwtToken,
			RefreshToken: refreshToken,
		}, nil
	}

	// Check if social account exists
	socialAccount, err := s.socialAccountRepo.GetByProviderAndSocialID("google", googleUser.ID)
	if err == nil {
		// Social account exists, get user and create auth response
		user, err := s.userRepo.GetUserByID(socialAccount.UserID)
		if err != nil {
			return nil, errors.New("user linked to social account not found")
		}
		return createAuthResponse(user)
	}

	// Social account doesn't exist, check if user with that email already exists
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
		return createAuthResponse(existingUser)
	}

	// No existing user or social account, create a new user
	newUser := &domain.User{
		Username:     s.generateUsernameFromEmail(googleUser.Email),
		Email:        googleUser.Email,
		PasswordHash: "", // No password for OAuth users
		FullName:     googleUser.Name,
		Role:         domain.UserRoleUser,
		IsActive:     true,
	}
	if err := s.userRepo.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("failed to create user from google auth: %w", err)
	}

	// Link the new social account
	socialAccount = &domain.SocialAccount{
		UserID:   newUser.ID,
		Provider: "google",
		SocialID: googleUser.ID,
		Email:    googleUser.Email,
	}
	if err := s.socialAccountRepo.Create(socialAccount); err != nil {
		// Attempt to clean up the newly created user if linking fails
		if err := s.userRepo.DeleteUser(newUser.ID); err != nil {
			// Log lại để dev biết hệ thống đang bị rác dữ liệu
			slog.Error("failed to cleanup user after social account link failure",
				"user_id", newUser.ID.String(),
				"error", err)
		}
		return nil, fmt.Errorf("failed to create social account link: %w", err)
	}

	return createAuthResponse(newUser)
}

// CreateSession creates a new session for the user
func (s *authService) CreateSession(userID uuid.UUID, userAgent, clientIP string) (*domain.Session, string, error) {
	// Generate refresh token
	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	session := &domain.Session{
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(s.refreshTokenExpiresIn),
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

// GetSessionByRefreshToken retrieves a session by its refresh token
func (s *authService) GetSessionByRefreshToken(token string) (*domain.Session, error) {
	return s.sessionRepo.GetByRefreshToken(token)
}

// getGoogleUserInfo fetches user info from Google API
func (s *authService) getGoogleUserInfo(accessToken string) (*dto.GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo dto.GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// generateToken creates a JWT token for the user
func (s *authService) generateToken(user *domain.User, sessionID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(s.jwtExpiresIn).Unix(),
		"iat":      time.Now().Unix(),
		"jti":      sessionID.String(), // JWT ID, linking the token to the session
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken generates a secure random refresh token
func (s *authService) generateRefreshToken() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
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
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// UpdateMe updates the currently authenticated user's profile
func (s *authService) UpdateMe(userID uuid.UUID, req dto.UpdateMeRequest) (*dto.UserResponse, error) {
	// Get current user
	currentUser, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	updates := make(map[string]interface{})

	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}

	if req.Email != nil && *req.Email != currentUser.Email {
		// Check if the new email is already taken
		if _, err := s.userRepo.GetUserByEmail(*req.Email); err == nil {
			return nil, errors.New("email is already in use")
		}
		updates["email"] = *req.Email
	}

	// If there are no updates, just return the current user
	if len(updates) == 0 {
		return s.toUserResponse(currentUser), nil
	}

	// Apply updates
	if err := s.userRepo.UpdateUser(userID, updates); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Fetch updated user to return
	updatedUser, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to fetch updated user information")
	}

	return s.toUserResponse(updatedUser), nil
}
