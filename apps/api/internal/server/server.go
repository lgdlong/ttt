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
	"api/internal/infrastructure"
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
	// Infrastructure layer
	openAIClient, err := infrastructure.NewOpenAIClient()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize OpenAI client - vector search will be disabled")
		openAIClient = nil // Continue without AI features
	}

	// Repository layer
	videoRepo := repository.NewVideoRepository(dbService.GetGormDB())
	userRepo := repository.NewUserRepository(dbService.GetGormDB())
	socialAccountRepo := repository.NewSocialAccountRepository(dbService.GetGormDB())
	sessionRepo := repository.NewSessionRepository(dbService.GetGormDB())
	tagRepo := repository.NewTagRepository(dbService.GetGormDB(), openAIClient)
	statsRepo := repository.NewStatsRepository(dbService.GetGormDB())
	reviewRepo := repository.NewVideoTranscriptReviewRepository(dbService.GetGormDB())

	// Service layer
	videoService := service.NewVideoService(videoRepo)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, socialAccountRepo, sessionRepo)
	tagService := service.NewTagService(tagRepo, videoRepo)
	tagServiceV2 := service.NewTagServiceV2(tagRepo, videoRepo)
	statsService := service.NewStatsService(statsRepo)
	reviewService := service.NewVideoTranscriptReviewService(reviewRepo, videoRepo, userRepo)

	// Handler layer
	videoHandler := handler.NewVideoHandler(videoService)
	searchHandler := handler.NewSearchHandler(videoService)
	systemHandler := handler.NewSystemHandler()
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	tagHandler := handler.NewTagHandler(tagService, tagServiceV2)
	statsHandler := handler.NewStatsHandler(statsService)
	reviewHandler := handler.NewVideoTranscriptReviewHandler(reviewService)

	// Register routes
	routes.RegisterRoutes(router, videoHandler, searchHandler, systemHandler, userHandler, authHandler, tagHandler, statsHandler, reviewHandler, userRepo)

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
