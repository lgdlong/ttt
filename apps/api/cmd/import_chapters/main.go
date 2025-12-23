package main

import (
	"api/internal/database"
	"api/internal/domain"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

type JSONOutput struct {
	Analysis struct {
		Summary string `json:"summary"`
	} `json:"analysis"`
	Transcript []struct {
		SegmentID int    `json:"segment_id"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		StartTime int    `json:"start_time"` // Optional: If AI adds it later
	} `json:"transcript"`
}

func main() {
	inputDir := flag.String("dir", "../../agents/resources/transcript_to_json/output", "Directory containing JSON files")
	flag.Parse()

	log.Println("🔌 Connecting to Database...")
	dbService := database.New()
	if dbService == nil {
		log.Fatal("Could not initialize database service")
	}
	db := dbService.GetGormDB()

	files, err := os.ReadDir(*inputDir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", *inputDir, err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(*inputDir, file.Name())
		log.Printf("Processing %s...", file.Name())

		if err := processFile(db, filePath, file.Name()); err != nil {
			log.Printf("❌ Error processing %s: %v", file.Name(), err)
		} else {
			log.Printf("✅ Successfully processed %s", file.Name())
		}
	}
}

func processFile(db *gorm.DB, filePath, fileName string) error {
	// 1. Extract YoutubeID from filename (exactly first 11 characters)
	// Example: "_2oihXPeUwQ_Title.json" -> "_2oihXPeUwQ"
	if len(fileName) < 11 {
		return fmt.Errorf("filename too short, expected at least 11 characters for YouTube ID")
	}
	youtubeID := fileName[:11]

	// 2. Read and Parse JSON
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}

	var data JSONOutput
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 3. Find Video
		var video domain.Video
		if err := tx.Where("youtube_id = ?", youtubeID).First(&video).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("⚠️ Video %s not found in DB, skipping...", youtubeID)
				return nil
			}
			return err
		}

		// 4. Update Summary
		video.Summary = data.Analysis.Summary
		if err := tx.Save(&video).Error; err != nil {
			return err
		}

		// 5. Replace Chapters
		// Delete old chapters
		if err := tx.Where("video_id = ?", video.ID).Delete(&domain.VideoChapter{}).Error; err != nil {
			return err
		}

		// Insert new chapters
		if len(data.Transcript) > 0 {
			chapters := make([]domain.VideoChapter, len(data.Transcript))
			for i, seg := range data.Transcript {
				chapters[i] = domain.VideoChapter{
					VideoID:      video.ID,
					Title:        seg.Title,
					Content:      seg.Content,
					StartTime:    seg.StartTime,
					ChapterOrder: i + 1, // Order starts at 1
				}
			}
			if err := tx.Create(&chapters).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
