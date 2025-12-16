package routes

import (
	"api/internal/domain"
	"api/internal/handler"
	"api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all API routes
func RegisterRoutes(
	router *gin.Engine,
	videoHandler *handler.VideoHandler,
	searchHandler *handler.SearchHandler,
	systemHandler *handler.SystemHandler,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	tagHandler *handler.TagHandler,
	statsHandler *handler.StatsHandler,
	reviewHandler *handler.VideoTranscriptReviewHandler,
	userRepo domain.UserRepository,
) {
	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())
	router.Use(gin.Recovery())

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// System endpoints
		v1.GET("/health", systemHandler.Health)

		// Authentication endpoints (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/logout", authHandler.Logout)
			// TODO: TEMPORARILY DISABLED - Session and Refresh Token
			// auth.POST("/refresh", authHandler.RefreshToken)

			// Google OAuth
			auth.GET("/google", authHandler.GoogleAuth)
			auth.GET("/google/callback", authHandler.GoogleCallback)

			// Protected auth endpoints
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(userRepo))
			{
				authProtected.GET("/me", authHandler.Me)
				authProtected.PATCH("/me", authHandler.UpdateMe)
				authProtected.GET("/sessions", authHandler.GetActiveSessions)
			}
		}

		// User endpoints - requires admin role
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(userRepo))
		users.Use(middleware.RequireAdmin())
		{
			users.GET("", userHandler.ListUsers)
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Video endpoints (public)
		videos := v1.Group("/videos")
		{
			videos.GET("", videoHandler.GetVideoList)
			videos.GET("/:id", videoHandler.GetVideoDetail)
			videos.GET("/:id/transcript", videoHandler.GetVideoTranscript)

			// Review endpoints (protected - requires authentication)
			videoReviews := videos.Group("/:id/reviews")
			videoReviews.Use(middleware.AuthMiddleware(userRepo))
			{
				videoReviews.POST("", reviewHandler.SubmitReview)                // Submit review
				videoReviews.GET("/stats", reviewHandler.GetVideoReviewStats)    // Get review count
				videoReviews.GET("/status", reviewHandler.CheckUserReviewStatus) // Check if user reviewed
			}
		}

		// Transcript segment endpoints (protected - for mods/admins)
		segments := v1.Group("/transcript-segments")
		segments.Use(middleware.AuthMiddleware(userRepo))
		segments.Use(middleware.RequireMod())
		{
			segments.PATCH("/:id", videoHandler.UpdateSegment)
		}

		// Search endpoints (public)
		search := v1.Group("/search")
		{
			search.GET("/transcript", searchHandler.SearchTranscript)
			search.GET("/tags", searchHandler.SearchTags)
		}

		// Tags endpoints (public - for tag navigation)
		tags := v1.Group("/tags")
		{
			tags.GET("", tagHandler.ListCanonicalTags)   // List all tags
			tags.GET("/:id", tagHandler.GetCanonicalTag) // Get tag by ID
		}

		// Admin endpoints - requires admin role
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(userRepo))
		admin.Use(middleware.RequireAdmin())
		{
			// Statistics
			admin.GET("/stats", statsHandler.GetAdminStats)
		}

		// Mod endpoints - requires mod or admin role
		mod := v1.Group("/mod")
		mod.Use(middleware.AuthMiddleware(userRepo))
		mod.Use(middleware.RequireMod())
		{
			// Statistics
			mod.GET("/stats", statsHandler.GetModStats)

			// Tag management (REMOVED - use /api/v2/mod/tags)
			// Legacy Tag V1 routes removed, all tag operations moved to V2 API

			// Video management
			modVideos := mod.Group("/videos")
			{
				modVideos.GET("", videoHandler.GetModVideoList)
				modVideos.GET("/search", videoHandler.SearchVideos)
				modVideos.GET("/preview/:id", videoHandler.PreviewVideo)
				modVideos.POST("", videoHandler.CreateVideo)
				modVideos.DELETE("/:id", videoHandler.DeleteVideo)
				// Legacy Tag V1 routes removed - use /api/v2/mod/videos/:id/tags
				modVideos.POST("/:id/transcript/segments", videoHandler.CreateSegment)
			}
		}
	}

	// API v2 group - Canonical-Alias Tag Architecture
	v2 := router.Group("/api/v2")
	{
		// Mod endpoints - requires mod or admin role
		mod := v2.Group("/mod")
		mod.Use(middleware.AuthMiddleware(userRepo))
		mod.Use(middleware.RequireMod())
		{
			// Canonical Tag management (v2)
			modTags := mod.Group("/tags")
			{
				modTags.GET("", tagHandler.ListCanonicalTags)               // List all canonical tags
				modTags.POST("", tagHandler.CreateCanonicalTag)             // Create with auto-resolution
				modTags.POST("/merge", tagHandler.MergeTags)                // Manually merge source into target
				modTags.GET("/search", tagHandler.SearchCanonicalTags)      // Search canonical tags
				modTags.GET("/:id", tagHandler.GetCanonicalTag)             // Get by ID
				modTags.PATCH("/:id/approve", tagHandler.UpdateTagApproval) // Update approval status
			}

			// Video-Tag management (v2 - uses canonical tags)
			modVideos := mod.Group("/videos")
			{
				modVideos.GET("/:id/tags", tagHandler.GetVideoCanonicalTags)                  // Get video's canonical tags
				modVideos.POST("/:id/tags", tagHandler.AddCanonicalTagToVideo)                // Add tag with auto-resolution
				modVideos.DELETE("/:id/tags/:tag_id", tagHandler.RemoveCanonicalTagFromVideo) // Remove canonical tag
			}
		}
	}
}
