package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"api/internal/domain"
	"api/internal/infrastructure"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MigrationScript migrates existing tags table to canonical-alias architecture
// Strategy: For each old tag, create 1 canonical + 1 initial alias
// Preserves UUIDs to avoid breaking video_tags relationships

func main() {
	fmt.Println("========================================")
	fmt.Println("Tag Migration: Single → Canonical-Alias")
	fmt.Println("========================================\n")

	// Load environment
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: .env file not found, using system environment")
	}

	// Connect to database
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("✓ Database connected\n")

	// Initialize OpenAI client for embedding generation
	openAIClient, err := infrastructure.NewOpenAIClient()
	if err != nil {
		log.Fatalf("Failed to initialize OpenAI client: %v", err)
	}
	fmt.Println("✓ OpenAI client initialized\n")

	// Run migration
	ctx := context.Background()
	if err := migrateTagsToCanonicalAlias(ctx, db, openAIClient); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("\n========================================")
	fmt.Println("Migration completed successfully!")
	fmt.Println("========================================")
}

func connectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrateTagsToCanonicalAlias(ctx context.Context, db *gorm.DB, openAIClient *infrastructure.OpenAIClient) error {
	// Step 1: Verify old tables exist
	fmt.Println("Step 1: Checking old schema...")
	if !db.Migrator().HasTable("tags") {
		return fmt.Errorf("old 'tags' table not found - nothing to migrate")
	}
	fmt.Println("✓ Old 'tags' table found")

	// Step 2: Create new tables
	fmt.Println("\nStep 2: Creating new tables...")
	if err := db.AutoMigrate(&domain.CanonicalTag{}, &domain.TagAlias{}); err != nil {
		return fmt.Errorf("failed to create new tables: %w", err)
	}
	fmt.Println("✓ New tables created (canonical_tags, tag_aliases)")

	// Step 3: Load all old tags
	fmt.Println("\nStep 3: Loading old tags...")
	var oldTags []domain.Tag
	if err := db.Find(&oldTags).Error; err != nil {
		return fmt.Errorf("failed to load old tags: %w", err)
	}
	fmt.Printf("✓ Loaded %d old tags\n", len(oldTags))

	if len(oldTags) == 0 {
		fmt.Println("⚠ No old tags to migrate")
		return nil
	}

	// Step 4: Migrate each tag
	fmt.Println("\nStep 4: Migrating tags to canonical-alias structure...")
	fmt.Println("----------------------------------------")

	successCount := 0
	failCount := 0

	for i, oldTag := range oldTags {
		fmt.Printf("\n[%d/%d] Migrating: '%s' (ID: %s)\n", i+1, len(oldTags), oldTag.Name, oldTag.ID)

		if err := migrateOneTag(ctx, db, openAIClient, &oldTag); err != nil {
			fmt.Printf("  ✗ FAILED: %v\n", err)
			failCount++
			continue
		}

		fmt.Printf("  ✓ SUCCESS\n")
		successCount++

		// Rate limiting: Wait 100ms between API calls to avoid hitting OpenAI limits
		if openAIClient != nil && i < len(oldTags)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Println("\n----------------------------------------")
	fmt.Printf("Migration Results:\n")
	fmt.Printf("  Success: %d\n", successCount)
	fmt.Printf("  Failed:  %d\n", failCount)

	// Step 5: Migrate video_tags relationships
	fmt.Println("\nStep 5: Migrating video_tags relationships...")
	if err := migrateVideoTags(db); err != nil {
		return fmt.Errorf("failed to migrate video_tags: %w", err)
	}
	fmt.Println("✓ video_tags migrated")

	return nil
}

func migrateOneTag(ctx context.Context, db *gorm.DB, openAIClient *infrastructure.OpenAIClient, oldTag *domain.Tag) error {
	// Use transaction for atomicity
	return db.Transaction(func(tx *gorm.DB) error {
		// Create canonical tag (preserve old UUID)
		canonical := &domain.CanonicalTag{
			ID:          oldTag.ID, // CRITICAL: Preserve UUID for video_tags FK
			Slug:        domain.GenerateSlug(oldTag.Name),
			DisplayName: oldTag.Name,
			CreatedAt:   oldTag.CreatedAt,
			UpdatedAt:   oldTag.UpdatedAt,
		}

		if err := tx.Create(canonical).Error; err != nil {
			return fmt.Errorf("failed to create canonical: %w", err)
		}
		fmt.Printf("  → Created canonical: '%s'\n", canonical.DisplayName)

		// Create initial alias
		alias := &domain.TagAlias{
			CanonicalTagID:  canonical.ID,
			RawText:         oldTag.Name,
			NormalizedText:  domain.NormalizeText(oldTag.Name),
			Embedding:       oldTag.Embedding,
			IsReviewed:      true, // Mark as reviewed (original data)
			SimilarityScore: 1.0,  // Perfect match (canonical)
			CreatedAt:       oldTag.CreatedAt,
		}

		// If old tag had no embedding, generate it now
		if len(alias.Embedding.Slice()) == 0 && openAIClient != nil {
			fmt.Printf("  → Generating embedding (old tag had none)...\n")
			embedding, err := openAIClient.GetEmbedding(ctx, oldTag.Name)
			if err != nil {
				fmt.Printf("  ⚠ Warning: Failed to generate embedding: %v\n", err)
				// Continue without embedding - can be backfilled later
			} else {
				alias.Embedding = embedding
				fmt.Printf("  → Embedding generated (%d dims)\n", len(embedding.Slice()))
			}
		}

		if err := tx.Create(alias).Error; err != nil {
			return fmt.Errorf("failed to create alias: %w", err)
		}
		fmt.Printf("  → Created alias: '%s' → '%s'\n", alias.RawText, canonical.DisplayName)

		return nil
	})
}

func migrateVideoTags(db *gorm.DB) error {
	// Check if old video_tags backup exists
	if !db.Migrator().HasTable("video_tags_old_backup") {
		fmt.Println("  ⚠ No video_tags_old_backup found, assuming fresh installation")
		return nil
	}

	// Copy relationships from backup to new table
	// tag_id (old) → canonical_tag_id (new) - but they have same UUID!
	result := db.Exec(`
		INSERT INTO video_tags (video_id, canonical_tag_id, created_at)
		SELECT video_id, tag_id, NOW()
		FROM video_tags_old_backup
		ON CONFLICT (video_id, canonical_tag_id) DO NOTHING
	`)

	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("  → Migrated %d video-tag relationships\n", result.RowsAffected)
	return nil
}
