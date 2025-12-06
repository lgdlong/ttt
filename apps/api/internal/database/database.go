package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// GetDB returns the underlying *sql.DB instance for raw queries
	GetDB() *sql.DB

	// GetGormDB returns the underlying *gorm.DB instance
	GetGormDB() *gorm.DB
}

type service struct {
	db     *sql.DB
	gormDB *gorm.DB
	dbName string
}

var (
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	// Load .env file from project root
	// Try multiple paths to handle different working directories
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
	}

	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("Loaded .env from: %s", envPath)
			break
		}
	}

	// Get env variables
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbUser := os.Getenv("DB_USER")
	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	// Default values
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	// Validate required env vars
	if dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Missing required database env vars: DB_USER, DB_PASSWORD, DB_NAME")
	}

	// Log connection info for debugging
	log.Printf("Connecting to database: %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)

	// PostgreSQL connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)

	// Initialize SQL connection
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Printf("Failed to open database connection: %v", err)
		log.Printf("Connection string: postgres://%s:***@%s:%s/%s?sslmode=%s", dbUser, dbHost, dbPort, dbName, dbSSLMode)
		log.Fatal(err)
	}

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
		log.Printf("Make sure PostgreSQL is running at %s:%s", dbHost, dbPort)
		log.Fatal(err)
	}

	// Initialize GORM
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("Failed to initialize GORM: %v", err)
		log.Fatal(err)
	}

	dbInstance = &service{
		db:     sqlDB,
		gormDB: gormDB,
		dbName: dbName,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", s.dbName)
	return s.db.Close()
}

// GetDB returns the underlying *sql.DB instance for raw queries
func (s *service) GetDB() *sql.DB {
	return s.db
}

// GetGormDB returns the underlying *gorm.DB instance
func (s *service) GetGormDB() *gorm.DB {
	return s.gormDB
}
