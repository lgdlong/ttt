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

	// Enable pgcrypto extension (REQUIRED for uuid functions)
	if err := gormDB.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		log.Printf("⚠ pgcrypto extension failed: %v", err)
		return fmt.Errorf("pgcrypto extension is required: %w", err)
	} else {
		log.Println("✓ pgcrypto extension enabled")
	}

	// Verify uuid_generate_v4 function exists, create if needed
	if err := gormDB.Exec(`
		CREATE OR REPLACE FUNCTION uuid_generate_v4()
		RETURNS uuid AS
		'pgcrypto'
		LANGUAGE C IMMUTABLE STRICT;
	`).Error; err != nil {
		log.Printf("⚠ Warning: Could not ensure uuid_generate_v4 exists: %v", err)
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

	// Clean up orphan transcript_segments before adding FK constraint
	cleanOrphanSQL := `
		DELETE FROM transcript_segments 
		WHERE video_id NOT IN (SELECT id FROM videos)
	`
	if err := gormDB.Exec(cleanOrphanSQL).Error; err != nil {
		log.Printf("⚠ Warning: Could not clean orphan transcript_segments: %v", err)
	} else {
		log.Println("✓ Orphan transcript_segments cleaned")
	}

	// Migrate TranscriptSegment (TSV field is ignored, will add manually)
	if err := gormDB.AutoMigrate(&domain.TranscriptSegment{}); err != nil {
		return fmt.Errorf("migration failed for TranscriptSegment: %w", err)
	}
	log.Println("✓ TranscriptSegment table migrated")

	// Clean up orphan video_transcript_reviews before adding FK constraints
	cleanReviewsSQL := `
		DELETE FROM video_transcript_reviews 
		WHERE video_id NOT IN (SELECT id FROM videos)
		   OR user_id NOT IN (SELECT id FROM users)
	`
	if err := gormDB.Exec(cleanReviewsSQL).Error; err != nil {
		log.Printf("⚠ Warning: Could not clean orphan video_transcript_reviews: %v", err)
	} else {
		log.Println("✓ Orphan video_transcript_reviews cleaned")
	}

        // Migrate VideoTranscriptReview
        if err := gormDB.AutoMigrate(&domain.VideoTranscriptReview{}); err != nil {
                return fmt.Errorf("migration failed for VideoTranscriptReview: %w", err)
        }
        log.Println("âœ“ VideoTranscriptReview table migrated")

        // Migrate VideoChapter (Semantic Chapters)
        if err := gormDB.AutoMigrate(&domain.VideoChapter{}); err != nil {
                return fmt.Errorf("migration failed for VideoChapter: %w", err)
        }
        log.Println("âœ“ VideoChapter table migrated")

        // Migrate CanonicalTag (new canonical-alias architecture)
	// First, drop old constraint if it exists (for idempotency)
	gormDB.Exec("ALTER TABLE IF EXISTS canonical_tags DROP CONSTRAINT IF EXISTS uni_canonical_tags_slug")

	if err := gormDB.AutoMigrate(&domain.CanonicalTag{}); err != nil {
		return fmt.Errorf("migration failed for CanonicalTag: %w", err)
	}
	log.Println("✓ CanonicalTag table migrated")

	// Migrate TagAlias (new canonical-alias architecture)
	if err := gormDB.AutoMigrate(&domain.TagAlias{}); err != nil {
		return fmt.Errorf("migration failed for TagAlias: %w", err)
	}
	log.Println("✓ TagAlias table migrated")

	// Migrate User model
	if err := gormDB.AutoMigrate(&domain.User{}); err != nil {
		return fmt.Errorf("migration failed for User: %w", err)
	}
	log.Println("✓ User table migrated")

	// Migrate SocialAccount model
	if err := gormDB.AutoMigrate(&domain.SocialAccount{}); err != nil {
		return fmt.Errorf("migration failed for SocialAccount: %w", err)
	}
	log.Println("✓ SocialAccount table migrated")

	// Migrate Session model
	if err := gormDB.AutoMigrate(&domain.Session{}); err != nil {
		return fmt.Errorf("migration failed for Session: %w", err)
	}
	log.Println("✓ Session table migrated")

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
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_social_accounts_provider_id ON social_accounts(provider, provider_id)",
		"CREATE INDEX IF NOT EXISTS idx_social_accounts_user_id ON social_accounts(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token)",
		// Canonical-Alias architecture indexes
		"CREATE INDEX IF NOT EXISTS idx_canonical_tags_slug ON canonical_tags(slug)",
		"CREATE INDEX IF NOT EXISTS idx_canonical_tags_display_name ON canonical_tags(display_name)",
		"CREATE INDEX IF NOT EXISTS idx_tag_aliases_canonical_tag_id ON tag_aliases(canonical_tag_id)",
		"CREATE INDEX IF NOT EXISTS idx_tag_aliases_normalized ON tag_aliases(normalized_text)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_aliases_normalized_unique ON tag_aliases(normalized_text)",
		"CREATE INDEX IF NOT EXISTS idx_video_canonical_tags_video_id ON video_canonical_tags(video_id)",
		"CREATE INDEX IF NOT EXISTS idx_video_canonical_tags_canonical_tag_id ON video_canonical_tags(canonical_tag_id)",
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

	// HNSW vector index for semantic search on tag_aliases.embedding (Canonical-Alias)
	hswVectorIndexSQL := `CREATE INDEX IF NOT EXISTS idx_tag_aliases_embedding_hnsw ON tag_aliases USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)`
	if err := db.Exec(hswVectorIndexSQL).Error; err != nil {
		log.Printf("Warning: failed to create HNSW vector index for tag_aliases (pgvector may not be installed): %v", err)
	} else {
		log.Println("  ✓ HNSW vector index for tag_aliases.embedding created")
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
                &domain.Session{},
                &domain.SocialAccount{},
                &domain.User{},
                &domain.TranscriptSegment{},
                &domain.VideoChapter{},
                &domain.Video{},
                &domain.VideoTranscriptReview{},
                &domain.TagAlias{},
                &domain.CanonicalTag{},
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
                "users":                    &domain.User{},
                "social_accounts":          &domain.SocialAccount{},
                "sessions":                 &domain.Session{},
                "videos":                   &domain.Video{},
                "transcript_segments":      &domain.TranscriptSegment{},
                "video_chapters":           &domain.VideoChapter{},
                "video_transcript_reviews": &domain.VideoTranscriptReview{},
                "canonical_tags":           &domain.CanonicalTag{},
                "tag_aliases":              &domain.TagAlias{},
        }

        for name, model := range models {		if gormDB.Migrator().HasTable(model) {
			log.Printf("✓ Table exists: %s", name)
		} else {
			log.Printf("✗ Table missing: %s", name)
		}
	}

	log.Println("\n==================================")
	return nil
}
