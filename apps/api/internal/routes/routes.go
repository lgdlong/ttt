package routes

import (
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

		// Video endpoints
		videos := v1.Group("/videos")
		{
			videos.GET("", videoHandler.GetVideoList)
			videos.GET("/:id", videoHandler.GetVideoDetail)
			videos.GET("/:id/transcript", videoHandler.GetVideoTranscript)
		}

		// Search endpoints
		search := v1.Group("/search")
		{
			search.GET("/transcript", searchHandler.SearchTranscript)
			search.GET("/tags", searchHandler.SearchTags)
		}
	}
}
