package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/pgvector/pgvector-go"
	openai "github.com/sashabaranov/go-openai"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Tag model - matches domain.Tag
type Tag struct {
	ID        string          `gorm:"type:uuid;primaryKey"`
	Name      string          `gorm:"type:varchar(100)"`
	Embedding pgvector.Vector `gorm:"type:vector(1536)"` // text-embedding-3-small native dimensions
}

func (Tag) TableName() string {
	return "tags"
}

// Hàm này sẽ:
// - Tìm tất cả các tag chưa có embedding (cột embedding IS NULL).
// - Gọi OpenAI API để tạo embedding vector cho từng tag (dùng model text-embedding-3-small).
// - Truncate embeddings to 2000 dimensions (PostgreSQL pgvector limit).
// - Lưu embedding vào database.
// - Báo cáo số lượng thành công/thất bại và tổng chi phí.
func main() {
	// Load .env
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: .env file not found, using system environment")
	}

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set in environment")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	// Find tags without embeddings
	var tagsWithoutEmbedding []Tag
	if err := db.Where("embedding IS NULL").Find(&tagsWithoutEmbedding).Error; err != nil {
		log.Fatalf("Failed to query tags: %v", err)
	}

	total := len(tagsWithoutEmbedding)
	if total == 0 {
		log.Println("✅ All tags already have embeddings!")
		return
	}

	log.Printf("📊 Found %d tags without embeddings\n", total)
	log.Println("🚀 Starting backfill process...")
	log.Println("")

	ctx := context.Background()
	successCount := 0
	failCount := 0
	totalCost := 0.0

	for i, tag := range tagsWithoutEmbedding {
		log.Printf("[%d/%d] Processing: %s (ID: %s)", i+1, total, tag.Name, tag.ID)

		// Call OpenAI API with text-embedding-3-small
		resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Input: []string{tag.Name},
			Model: openai.SmallEmbedding3, // text-embedding-3-small (1536 dims)
		})

		if err != nil {
			log.Printf("  ❌ Failed: %v\n", err)
			failCount++

			// Rate limit handling
			if err.Error() == "rate limit exceeded" {
				log.Println("  ⏸️  Rate limit hit, waiting 60 seconds...")
				time.Sleep(60 * time.Second)
				i-- // Retry this tag
			}
			continue
		}

		// No truncation needed - text-embedding-3-small returns 1536 dimensions (native fit)
		tag.Embedding = pgvector.NewVector(resp.Data[0].Embedding)

		// Update tag with embedding
		if err := db.Save(&tag).Error; err != nil {
			log.Printf("  ❌ Failed to save: %v\n", err)
			failCount++
			continue
		}

		// Cost calculation (text-embedding-3-small: $0.00002 per 1K tokens)
		// Tag name ~ 1-3 tokens
		cost := 0.000002
		totalCost += cost

		log.Printf("  ✅ Success (cost: $%.6f)\n", cost)
		successCount++

		// Rate limiting: 3000 requests per minute for OpenAI
		// Sleep 20ms between requests = 50 req/sec = safe
		time.Sleep(20 * time.Millisecond)
	}

	// Summary
	log.Println("")
	log.Println("========================================")
	log.Println("📈 Backfill Complete!")
	log.Println("========================================")
	log.Printf("✅ Success: %d tags\n", successCount)
	log.Printf("❌ Failed:  %d tags\n", failCount)
	log.Printf("💰 Total cost: $%.6f\n", totalCost)
	log.Println("")

	if failCount > 0 {
		log.Println("⚠️  Some tags failed. Run the script again to retry.")
	} else {
		log.Println("🎉 All tags now have embeddings!")
	}
}
