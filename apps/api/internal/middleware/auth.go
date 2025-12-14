package middleware

import (
	"api/internal/domain"
	"api/internal/dto"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(userRepo domain.UserRepository) gin.HandlerFunc {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from cookie first
		if cookie, err := c.Cookie("token"); err == nil {
			tokenString = cookie
		}

		// Fall back to Authorization header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "No token provided",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid token",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid token claims",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Get user ID from claims
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid user ID in token",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid user ID format",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Get user from database to ensure they still exist and are active
		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not found",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Check if user is active
		if !user.IsActive {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Account is deactivated",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("email", user.Email)
		c.Set("role", user.Role)
		c.Set("user", &dto.UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})

		c.Next()
	}
}

// RequireRole creates middleware that requires specific role(s)
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Role not found in context",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		userRole := role.(string)

		// Check if user's role is in allowed roles
		allowed := false
		for _, r := range allowedRoles {
			if userRole == r {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Insufficient permissions",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a shorthand for requiring admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(string(domain.UserRoleAdmin))
}

// RequireMod is a shorthand for requiring mod or admin role
func RequireMod() gin.HandlerFunc {
	return RequireRole(string(domain.UserRoleAdmin), string(domain.UserRoleMod))
}

// OptionalAuth middleware allows both authenticated and unauthenticated access
// If a valid token is present, user info is set in context
func OptionalAuth(userRepo domain.UserRepository) gin.HandlerFunc {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from cookie first
		if cookie, err := c.Cookie("token"); err == nil {
			tokenString = cookie
		}

		// Fall back to Authorization header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// If no token, continue without authentication
		if tokenString == "" {
			c.Next()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		// If token is invalid, continue without authentication
		if err != nil || !token.Valid {
			c.Next()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		// Get user ID from claims
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.Next()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.Next()
			return
		}

		// Get user from database
		user, err := userRepo.GetUserByID(userID)
		if err != nil || !user.IsActive {
			c.Next()
			return
		}

		// Set user info in context
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("email", user.Email)
		c.Set("role", user.Role)
		c.Set("user", &dto.UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})

		c.Next()
	}
}
