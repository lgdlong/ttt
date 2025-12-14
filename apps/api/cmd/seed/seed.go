package main

import (
	"api/internal/domain"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Helper struct để parse JSON (vì JSON date là string, còn DB là time.Time)
type JsonVideo struct {
	YoutubeID    string `json:"youtube_id"`
	Title        string `json:"title"`
	PublishedAt  string `json:"published_at"`
	Duration     int    `json:"duration"`
	ViewCount    int    `json:"view_count"`
	ThumbnailURL string `json:"thumbnail_url"`
}

func ImportVideos(db *gorm.DB, jsonPath string) error {
	log.Println("--- BẮT ĐẦU IMPORT VIDEOS ---")

	// 1. Đọc file JSON
	file, err := os.Open(jsonPath)
	if err != nil {
		return fmt.Errorf("không thể mở file json: %w", err)
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	var jsonList []JsonVideo
	if err := json.Unmarshal(byteValue, &jsonList); err != nil {
		return fmt.Errorf("lỗi parse json: %w", err)
	}

	// 2. Convert sang Entity Domain
	var videos []domain.Video
	for _, j := range jsonList {
		// Parse ngày tháng: YYYY-MM-DD
		t, err := time.Parse("2006-01-02", j.PublishedAt)
		if err != nil {
			// Fallback nếu ngày lỗi hoặc rỗng
			t = time.Now()
		}

		// Thay thế default.jpg thành hqdefault.jpg để lấy thumbnail chất lượng cao hơn
		thumbnailURL := strings.Replace(j.ThumbnailURL, "default.jpg", "hqdefault.jpg", 1)

		videos = append(videos, domain.Video{
			YoutubeID:    j.YoutubeID,
			Title:        j.Title,
			PublishedAt:  t,
			Duration:     j.Duration,
			ViewCount:    j.ViewCount,
			ThumbnailURL: thumbnailURL,
		})
	}

	// 3. Batch Insert (100 video / lần)
	// Sử dụng OnConflict để nếu chạy lại thì không bị lỗi trùng lặp (Update lại thông tin)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "youtube_id"}}, // Khớp theo YoutubeID
			DoUpdates: clause.AssignmentColumns([]string{"title", "view_count", "duration", "published_at", "thumbnail_url"}),
		}).CreateInBatches(videos, 100).Error; err != nil {
			return err
		}
		log.Printf("✅ Đã import thành công %d videos", len(videos))
		return nil
	})
}

func ImportTranscripts(db *gorm.DB, tsvDir string) error {
	log.Println("--- BẮT ĐẦU IMPORT TRANSCRIPTS ---")

	// 1. Lấy Map [YoutubeID] -> [UUID] từ DB
	// Bước này cực quan trọng để tránh phải query DB trong vòng lặp đọc file
	log.Println("⏳ Đang load danh sách Video ID...")
	type VideoMap struct {
		ID        uuid.UUID
		YoutubeID string
	}
	var vidList []VideoMap
	if err := db.Model(&domain.Video{}).Select("id, youtube_id").Scan(&vidList).Error; err != nil {
		return fmt.Errorf("lỗi load video map: %w", err)
	}

	// Tạo Hash Map để tra cứu nhanh O(1)
	vidMap := make(map[string]uuid.UUID)
	for _, v := range vidList {
		vidMap[v.YoutubeID] = v.ID
	}
	log.Printf("... Đã load %d video từ DB.", len(vidMap))

	// 2. Quét file TSV
	files, err := filepath.Glob(filepath.Join(tsvDir, "*.tsv"))
	if err != nil {
		return err
	}

	var totalSegments int64 = 0
	var batchSegments []domain.TranscriptSegment
	const BatchSize = 2000 // Insert mỗi lần 2000 dòng

	for i, filePath := range files {
		// Lấy YoutubeID từ tên file (vd: "tsv_files/dQw4w9WgXcQ.tsv" -> "dQw4w9WgXcQ")
		fileName := filepath.Base(filePath)
		youtubeID := strings.TrimSuffix(fileName, ".tsv")

		// Tìm UUID tương ứng
		videoUUID, exists := vidMap[youtubeID]
		if !exists {
			log.Printf("⚠️  Bỏ qua file %s: Không tìm thấy Video ID trong DB", fileName)
			continue
		}

		// Đọc file TSV
		f, err := os.Open(filePath)
		if err != nil {
			log.Printf("❌ Lỗi mở file %s: %v", fileName, err)
			continue
		}

		reader := csv.NewReader(f)
		reader.Comma = '\t'      // Ngăn cách bằng Tab
		reader.LazyQuotes = true // Cho phép ký tự lạ

		// Bỏ qua header (nếu dòng đầu tiên là text không phải số)
		// Ta sẽ xử lý trong vòng lặp bên dưới

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}

			// record[0]: start, record[1]: end, record[2]: text
			if len(record) < 3 {
				continue
			}

			// Parse thời gian (giả sử file TSV chứa mili-giây dạng số nguyên)
			// Nếu file TSV của bạn là dạng "00:00:01,500" thì cần hàm convert riêng
			start, err1 := strconv.Atoi(record[0])
			end, err2 := strconv.Atoi(record[1])

			// Nếu dòng đầu là Header "start", "end" -> Atoi sẽ lỗi -> Bỏ qua dòng này
			if err1 != nil || err2 != nil {
				continue
			}

			batchSegments = append(batchSegments, domain.TranscriptSegment{
				VideoID:     videoUUID,
				StartTime:   start,
				EndTime:     end,
				TextContent: strings.TrimSpace(record[2]),
			})
		}
		f.Close()

		// Kiểm tra batch để insert
		if len(batchSegments) >= BatchSize {
			if err := db.CreateInBatches(batchSegments, len(batchSegments)).Error; err != nil {
				log.Printf("❌ Lỗi insert batch: %v", err)
			}
			totalSegments += int64(len(batchSegments))
			batchSegments = nil // Reset batch (giữ nguyên capacity để tối ưu mem)
			fmt.Printf(".")     // Tiến trình
		}

		if (i+1)%50 == 0 {
			fmt.Printf("\nĐã xử lý %d/%d files...", i+1, len(files))
		}
	}

	// Insert nốt những dòng còn sót lại trong batch cuối cùng
	if len(batchSegments) > 0 {
		if err := db.CreateInBatches(batchSegments, len(batchSegments)).Error; err != nil {
			log.Printf("❌ Lỗi insert batch cuối: %v", err)
		}
		totalSegments += int64(len(batchSegments))
	}

	log.Printf("\n✅ HOÀN TẤT! Đã import tổng cộng %d dòng sub.", totalSegments)
	return nil
}
