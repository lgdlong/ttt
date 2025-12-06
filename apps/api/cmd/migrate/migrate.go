package migrate

import (
	"api/internal/database"
	"api/internal/domain"
	"fmt"
	"log"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

// Migrate runs all database migrations using GORM AutoMigrate
func Migrate() error {
	db := database.New()

	// Check if db service is nil
	if db == nil {
		return fmt.Errorf("database initialization failed: database.New() returned nil service")
	}

	gormDB := db.GetGormDB()

	// Check if GORM instance is nil
	if gormDB == nil {
		return fmt.Errorf("database initialization failed: GORM instance is nil")
	}

	log.Println("Running migrations...")

	// Try to enable pgvector extension (optional - ignore if not available)
	if err := gormDB.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		log.Printf("⚠ pgvector extension not available (optional): %v", err)
	} else {
		log.Println("✓ pgvector extension enabled")
	}

	// AutoMigrate all models
	if err := gormDB.AutoMigrate(
		&domain.Video{},
		&domain.TranscriptSegment{},
		&domain.Tag{},
	).Error; err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Println("✓ All models migrated successfully")

	// Create necessary indexes for performance
	if err := createIndexes(gormDB); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	log.Println("✓ Indexes created")

	// Enable Full Text Search trigger for transcript segments
	if err := enableFTS(gormDB); err != nil {
		return fmt.Errorf("failed to enable full text search: %w", err)
	}
	log.Println("✓ Full text search enabled")

	log.Println("All migrations completed successfully!")
	return nil
}

// createIndexes creates additional indexes for performance
func createIndexes(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("gorm.DB is nil in createIndexes")
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_transcript_segments_video_id_start_time ON transcript_segments(video_id, start_time)",
		"CREATE INDEX IF NOT EXISTS idx_videos_youtube_id ON videos(youtube_id)",
		"CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name)",
	}

	for _, idx := range indexes {
		result := db.Exec(idx)
		if result.Error != nil {
			log.Printf("Warning: failed to create index '%s': %v", idx, result.Error)
			// Continue even if index creation fails
			continue
		}
	}

	return nil
}

// enableFTS enables Full Text Search for transcript segments
func enableFTS(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("gorm.DB is nil in enableFTS")
	}

	// Create trigger function for automatic TSV update
	triggerSQL := `
	CREATE OR REPLACE FUNCTION transcript_segments_tsvector_update()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.tsv := to_tsvector('simple', NEW.text_content);
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Drop trigger if exists and recreate
	DROP TRIGGER IF EXISTS transcript_segments_tsv_trigger ON transcript_segments;
	CREATE TRIGGER transcript_segments_tsv_trigger
	BEFORE INSERT OR UPDATE ON transcript_segments
	FOR EACH ROW
	EXECUTE FUNCTION transcript_segments_tsvector_update();
	`

	if err := db.Exec(triggerSQL).Error; err != nil {
		log.Printf("Warning: failed to enable full text search: %v", err)
		// Don't fail if FTS setup fails
		return nil
	}

	return nil
}

// Rollback rolls back all migrations (drops all tables)
// WARNING: This is destructive and only for development
func Rollback() error {
	db := database.New()
	gormDB := db.GetGormDB()

	log.Println("Rolling back all migrations...")

	// Drop all tables in reverse order of dependencies
	if err := gormDB.Migrator().DropTable(
		&domain.TranscriptSegment{},
		&domain.Video{},
		&domain.Tag{},
	).Error; err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	log.Println("All tables dropped successfully!")
	return nil
}

// Status prints the current migration status
func Status() error {
	db := database.New()
	gormDB := db.GetGormDB()

	log.Println("\n=== Database Migration Status ===")

	models := map[string]interface{}{
		"videos":              &domain.Video{},
		"transcript_segments": &domain.TranscriptSegment{},
		"tags":                &domain.Tag{},
	}

	for name, model := range models {
		if gormDB.Migrator().HasTable(model) {
			log.Printf("✓ Table exists: %s", name)
		} else {
			log.Printf("✗ Table missing: %s", name)
		}
	}

	log.Println("\n==================================")
	return nil
}
