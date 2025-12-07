package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"api/internal/database"
	"api/internal/handler"
	"api/internal/repository"
	"api/internal/routes"
	"api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	port int

	db database.Service
}

func init() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialize database
	dbService := database.New()

	// Set Gin mode based on environment
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Initialize dependency injection chain
	// Repository layer
	videoRepo := repository.NewVideoRepository(dbService.GetGormDB())

	// Service layer
	videoService := service.NewVideoService(videoRepo)

	// Handler layer
	videoHandler := handler.NewVideoHandler(videoService)
	searchHandler := handler.NewSearchHandler(videoService)
	systemHandler := handler.NewSystemHandler()

	// Register routes
	routes.RegisterRoutes(router, videoHandler, searchHandler, systemHandler)

	// Register Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Info().Msgf("Server starting on port %d", port)
	return server
}
