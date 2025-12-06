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

	// Enable pgcrypto extension (REQUIRED for gen_random_uuid())
	if err := gormDB.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		log.Printf("⚠ pgcrypto extension failed: %v", err)
	} else {
		log.Println("✓ pgcrypto extension enabled")
	}

	// Enable pgvector extension (REQUIRED for Tag.Embedding field)
	if err := gormDB.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		log.Printf("⚠ pgvector extension not available: %v", err)
		log.Println("⚠ Tag table with Embedding field may fail to migrate")
	} else {
		log.Println("✓ pgvector extension enabled")
	}

	// AutoMigrate Video first (no special types)
	if err := gormDB.AutoMigrate(&domain.Video{}); err != nil {
		return fmt.Errorf("migration failed for Video: %w", err)
	}
	log.Println("✓ Video table migrated")

	// Migrate TranscriptSegment (TSV field is ignored, will add manually)
	if err := gormDB.AutoMigrate(&domain.TranscriptSegment{}); err != nil {
		return fmt.Errorf("migration failed for TranscriptSegment: %w", err)
	}
	log.Println("✓ TranscriptSegment table migrated")

	// Migrate Tag (requires pgvector extension for Embedding field)
	if err := gormDB.AutoMigrate(&domain.Tag{}); err != nil {
		log.Printf("⚠ Warning: Tag migration failed (pgvector may not be installed): %v", err)
		// Try to create Tag table without Embedding column
		createTagSQL := `
			CREATE TABLE IF NOT EXISTS tags (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(100) UNIQUE NOT NULL
			)
		`
		if err := gormDB.Exec(createTagSQL).Error; err != nil {
			return fmt.Errorf("failed to create tags table: %w", err)
		}
		log.Println("✓ Tag table created (without Embedding column)")
	} else {
		log.Println("✓ Tag table migrated")
	}

	log.Println("✓ All models migrated successfully")

	// Add TSV column manually using raw SQL (GORM ignores it with gorm:"-")
	if err := addTSVColumn(gormDB); err != nil {
		log.Printf("⚠ Warning: could not add TSV column: %v", err)
	} else {
		log.Println("✓ TSV column added successfully")
	}

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

// addTSVColumn adds the TSV (tsvector) column to transcript_segments table using raw SQL
func addTSVColumn(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("gorm.DB is nil in addTSVColumn")
	}

	// Check if column already exists using raw SQL
	var count int64
	checkSQL := `
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_name = 'transcript_segments' 
		AND column_name = 'tsv'
		AND table_schema = CURRENT_SCHEMA()
	`
	if err := db.Raw(checkSQL).Scan(&count).Error; err != nil {
		return fmt.Errorf("failed to check TSV column existence: %w", err)
	}

	if count > 0 {
		log.Println("TSV column already exists")
		return nil
	}

	// Add TSV column using raw SQL
	addColumnSQL := `ALTER TABLE transcript_segments ADD COLUMN IF NOT EXISTS tsv tsvector`
	if err := db.Exec(addColumnSQL).Error; err != nil {
		return fmt.Errorf("failed to add TSV column: %w", err)
	}

	// Create GIN index on TSV column for full text search
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_transcript_segments_tsv ON transcript_segments USING gin(tsv)`
	if err := db.Exec(createIndexSQL).Error; err != nil {
		return fmt.Errorf("failed to create TSV index: %w", err)
	}

	return nil
}

// createIndexes creates additional indexes for performance
func createIndexes(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("gorm.DB is nil in createIndexes")
	}

	// Standard B-tree indexes
	btreeIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_transcript_segments_video_id_start_time ON transcript_segments(video_id, start_time)",
		"CREATE INDEX IF NOT EXISTS idx_videos_youtube_id ON videos(youtube_id)",
		"CREATE INDEX IF NOT EXISTS idx_videos_published_at ON videos(published_at)",
		"CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name)",
	}

	for _, idx := range btreeIndexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("Warning: failed to create index: %v", err)
			continue
		}
	}

	// GIN index for Full Text Search on transcript_segments.tsv (REQUIRED)
	ginIndexSQL := `CREATE INDEX IF NOT EXISTS idx_transcript_segments_tsv ON transcript_segments USING gin(tsv)`
	if err := db.Exec(ginIndexSQL).Error; err != nil {
		log.Printf("Warning: failed to create GIN index for FTS: %v", err)
	} else {
		log.Println("  ✓ GIN index for Full Text Search created")
	}

	// IVFFlat index for vector similarity search on tags.embedding (RECOMMENDED)
	// IVFFlat is faster for approximate nearest neighbor search
	// lists = sqrt(number of rows), start with 100 for small datasets
	vectorIndexSQL := `CREATE INDEX IF NOT EXISTS idx_tags_embedding ON tags USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)`
	if err := db.Exec(vectorIndexSQL).Error; err != nil {
		log.Printf("Warning: failed to create vector index (pgvector may not be installed): %v", err)
	} else {
		log.Println("  ✓ IVFFlat vector index for tags.embedding created")
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
		NEW.tsv := to_tsvector('english', NEW.text_content);
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;
	`

	if err := db.Exec(triggerSQL).Error; err != nil {
		log.Printf("Warning: failed to create trigger function: %v", err)
	}

	// Drop trigger if exists and recreate
	dropTriggerSQL := `DROP TRIGGER IF EXISTS transcript_segments_tsv_trigger ON transcript_segments;`
	if err := db.Exec(dropTriggerSQL).Error; err != nil {
		log.Printf("Warning: failed to drop existing trigger: %v", err)
	}

	// Create trigger
	createTriggerSQL := `
	CREATE TRIGGER transcript_segments_tsv_trigger
	BEFORE INSERT OR UPDATE ON transcript_segments
	FOR EACH ROW
	EXECUTE FUNCTION transcript_segments_tsvector_update();
	`

	if err := db.Exec(createTriggerSQL).Error; err != nil {
		log.Printf("Warning: failed to create trigger: %v", err)
		// Don't fail if trigger setup fails
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
