package routes

import (
	"api/internal/handler"
	"api/internal/middleware"
	"api/internal/repository"

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
	userRepo repository.UserRepository,
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
		}

		// Search endpoints (public)
		search := v1.Group("/search")
		{
			search.GET("/transcript", searchHandler.SearchTranscript)
			search.GET("/tags", searchHandler.SearchTags)
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

			// Tag management
			modTags := mod.Group("/tags")
			{
				modTags.GET("", tagHandler.ListTags)
				modTags.POST("", tagHandler.CreateTag)
				modTags.GET("/:id", tagHandler.GetTag)
				modTags.PUT("/:id", tagHandler.UpdateTag)
				modTags.DELETE("/:id", tagHandler.DeleteTag)
			}

			// Video management
			modVideos := mod.Group("/videos")
			{
				modVideos.GET("", videoHandler.GetModVideoList)
				modVideos.GET("/search", videoHandler.SearchVideos)
				modVideos.GET("/preview/:id", videoHandler.PreviewVideo)
				modVideos.POST("", videoHandler.CreateVideo)
				modVideos.DELETE("/:id", videoHandler.DeleteVideo)
				modVideos.GET("/:id/tags", tagHandler.GetVideoTags)
				modVideos.POST("/:id/tags", tagHandler.AddTagToVideo)
				modVideos.DELETE("/:id/tags/:tag_id", tagHandler.RemoveTagFromVideo)
			}
		}
	}
}
