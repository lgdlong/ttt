package handler

import (
	"api/internal/domain"
	"api/internal/dto"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// setAuthCookie sets the JWT token as an HTTP-only cookie
func (h *AuthHandler) setAuthCookie(c *gin.Context, token string) {
	// Check if we're in production
	secure := os.Getenv("ENV") == "production"

	c.SetCookie(
		"token",    // name
		token,      // value
		60*60*24*7, // maxAge: 7 days in seconds
		"/",        // path
		"",         // domain (empty = current domain)
		secure,     // secure (HTTPS only in production)
		true,       // httpOnly
	)
}

// clearAuthCookie removes the auth cookie
func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
	c.SetCookie(
		"token",
		"",
		-1, // maxAge: negative = delete cookie
		"/",
		"",
		false,
		true,
	)
}

// setRefreshCookie sets the refresh token as an HTTP-only cookie
func (h *AuthHandler) setRefreshCookie(c *gin.Context, refreshToken string) {
	secure := os.Getenv("ENV") == "production"

	c.SetCookie(
		"refresh_token", // name
		refreshToken,    // value
		60*60*24*30,     // maxAge: 30 days in seconds
		"/",             // path
		"",              // domain
		secure,          // secure
		true,            // httpOnly
	)
}

// clearRefreshCookie removes the refresh token cookie
func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)
}

// Login godoc
// @Summary User login
// @Description Authenticate user with username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	response, err := h.service.Login(req, userAgent, clientIP)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if err.Error() == "account is deactivated" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Authentication failed",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	// Set token in cookie
	h.setAuthCookie(c, response.Token)
	h.setRefreshCookie(c, response.RefreshToken)

	c.JSON(http.StatusOK, response)
}

// Signup godoc
// @Summary User signup
// @Description Register a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body dto.SignupRequest true "Signup information"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /auth/signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	response, err := h.service.Signup(req, userAgent, clientIP)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Signup failed",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	// Set token in cookie
	h.setAuthCookie(c, response.Token)
	h.setRefreshCookie(c, response.RefreshToken)

	c.JSON(http.StatusCreated, response)
}

// Logout godoc
// @Summary User logout
// @Description Logout the current user and invalidate the session
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Invalidate the session on the server side
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && refreshToken != "" {
		// Find session by refresh token
		session, err := h.service.GetSessionByRefreshToken(refreshToken)
		if err == nil && session != nil {
			// Invalidate the specific session
			_ = h.service.Logout(session.ID) // We can ignore the error here
		}
	}

	// Clear cookies on the client side regardless
	h.clearAuthCookie(c)
	h.clearRefreshCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// LogoutAll godoc
// @Summary User logout from all devices
// @Description Logout the current user from all devices and invalidate all their sessions
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// Get user ID from context
	userIDCtx, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, ok := userIDCtx.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Invalid user ID format in context",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Invalidate all user's sessions on the server side
	_ = h.service.LogoutAll(userID) // We can ignore the error here

	// Clear cookies on the client side regardless
	h.clearAuthCookie(c)
	h.clearRefreshCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out from all devices successfully",
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using refresh token from cookie
// @Tags Authentication
// @Produce json
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Refresh token not found",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	response, err := h.service.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Set new access token in cookie
	h.setAuthCookie(c, response.Token)

	c.JSON(http.StatusOK, response)
}

// GoogleAuth godoc
// @Summary Initiate Google OAuth login
// @Description Redirect to Google OAuth consent page
// @Tags Authentication
// @Produce json
// @Success 200 {object} dto.GoogleAuthURLResponse
// @Router /auth/google [get]
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	// Generate state token for CSRF protection
	state := h.service.GenerateStateToken()

	// Store state in cookie for validation
	secure := os.Getenv("ENV") == "production"
	c.SetCookie(
		"oauth_state",
		state,
		60*10, // 10 minutes
		"/",
		"",
		secure,
		true,
	)

	// Get Google OAuth URL
	url := h.service.GetGoogleAuthURL(state)

	c.JSON(http.StatusOK, dto.GoogleAuthURLResponse{
		URL: url,
	})
}

// GoogleCallback godoc
// @Summary Handle Google OAuth callback
// @Description Process Google OAuth callback and authenticate user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State token for CSRF protection"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// Validate state token
	savedState, err := c.Cookie("oauth_state")
	if err != nil || savedState != state {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: "Invalid state token",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Get user agent and client IP
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	// Handle callback
	response, err := h.service.HandleGoogleCallback(code, userAgent, clientIP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication failed",
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Set token in cookie
	h.setAuthCookie(c, response.Token)

	// Redirect to frontend with success
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// Redirect based on role
	redirectPath := "/"
	if response.User.Role == "admin" {
		redirectPath = "/admin"
	} else if response.User.Role == "mod" {
		redirectPath = "/mod"
	}

	c.Redirect(http.StatusTemporaryRedirect, frontendURL+redirectPath)
}

// Me godoc
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userResponse, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not found in context",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

// UpdateMe godoc
// @Summary Update current user
// @Description Update the currently authenticated user's profile information (full_name, email)
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body dto.UpdateMeRequest true "Profile information to update"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or validation error"
// @Failure 401 {object} dto.ErrorResponse "User not authenticated"
// @Failure 409 {object} dto.ErrorResponse "Email already in use"
// @Router /auth/me [patch]
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not found in context",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// We know user is a *dto.UserResponse from the middleware
	currentUser := userCtx.(*dto.UserResponse)
	userID, err := uuid.Parse(currentUser.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Invalid user ID in context",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Bind request body
	var req dto.UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service to update the user
	updatedUser, err := h.service.UpdateMe(userID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "email is already in use" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to update profile",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// GetActiveSessions godoc
// @Summary Get active sessions
// @Description Get all active sessions for the current user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.SessionResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/sessions [get]
func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	// This would require session repository access
	// For now, return a placeholder response
	c.JSON(http.StatusOK, []dto.SessionResponse{
		{
			ID:        "current",
			UserAgent: c.GetHeader("User-Agent"),
			ClientIP:  c.ClientIP(),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
			IsBlocked: false,
		},
	})
}
