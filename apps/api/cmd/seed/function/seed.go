package function

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
	log.Printf("--- BẮT ĐẦU IMPORT VIDEOS TỪ: %s ---", jsonPath)

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

	if len(jsonList) == 0 {
		log.Println("⚠️ File JSON rỗng, không có video nào để import.")
		return nil
	}

	// 2. Convert sang Entity Domain
	var videos []domain.Video
	for _, j := range jsonList {
		// Parse ngày tháng: YYYY-MM-DD
		t, err := time.Parse("2006-01-02", j.PublishedAt)
		if err != nil {
			t = time.Now()
		}

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
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "youtube_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "view_count", "duration", "published_at", "thumbnail_url"}),
		}).CreateInBatches(videos, 100).Error; err != nil {
			return err
		}
		log.Printf("✅ Đã import/cập nhật thành công %d videos", len(videos))
		return nil
	})
}

func ImportTranscripts(db *gorm.DB, tsvDir string, force bool) error {
	log.Printf("--- BẮT ĐẦU IMPORT TRANSCRIPTS TỪ: %s (force=%v) ---", tsvDir, force)

	// 1. Lấy Map [YoutubeID] -> {UUID, HasTranscript}
	log.Println("⏳ Đang load danh sách Video từ DB...")
	type VideoInfo struct {
		ID            uuid.UUID
		YoutubeID     string
		HasTranscript bool
	}
	var vidList []VideoInfo
	if err := db.Model(&domain.Video{}).Select("id, youtube_id, has_transcript").Scan(&vidList).Error; err != nil {
		return fmt.Errorf("lỗi load video map: %w", err)
	}

	vidMap := make(map[string]VideoInfo)
	for _, v := range vidList {
		vidMap[v.YoutubeID] = v
	}
	log.Printf("... Đã load %d video từ DB.", len(vidMap))

	// 2. Quét file TSV
	files, err := filepath.Glob(filepath.Join(tsvDir, "*.tsv"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		log.Println("⚠️ Không tìm thấy file TSV nào.")
		return nil
	}

	var stats struct {
		Success int
		Skipped int
		Failed  int
		Total   int
	}
	stats.Total = len(files)

	for i, filePath := range files {
		fileName := filepath.Base(filePath)
		youtubeID := strings.TrimSuffix(fileName, ".tsv")

		vInfo, exists := vidMap[youtubeID]
		if !exists {
			log.Printf("[%d/%d] ⚠️ Bỏ qua %s: Không có trong DB", i+1, stats.Total, fileName)
			stats.Skipped++
			continue
		}

		// Kiểm tra nếu đã có transcript và không ép buộc (force)
		if vInfo.HasTranscript && !force {
			stats.Skipped++
			continue
		}

		// Xử lý từng file trong một Transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			// Nếu force, xóa transcript cũ trước
			if vInfo.HasTranscript && force {
				if err := tx.Where("video_id = ?", vInfo.ID).Delete(&domain.TranscriptSegment{}).Error; err != nil {
					return fmt.Errorf("lỗi xóa transcript cũ: %w", err)
				}
			}

			// Đọc và parse file TSV
			segments, err := parseTSV(filePath, vInfo.ID)
			if err != nil {
				return err
			}

			if len(segments) == 0 {
				return fmt.Errorf("file rỗng hoặc không có dữ liệu hợp lệ")
			}

			// Insert segments theo batch nhỏ trong transaction của file này
			if err := tx.CreateInBatches(segments, 1000).Error; err != nil {
				return fmt.Errorf("lỗi insert segments: %w", err)
			}

			// Cập nhật trạng thái cho Video
			if err := tx.Model(&domain.Video{}).Where("id = ?", vInfo.ID).Update("has_transcript", true).Error; err != nil {
				return fmt.Errorf("lỗi cập nhật has_transcript: %w", err)
			}

			return nil
		})

		if err != nil {
			log.Printf("[%d/%d] ❌ Lỗi file %s: %v", i+1, stats.Total, fileName, err)
			stats.Failed++
		} else {
			stats.Success++
			if stats.Success%50 == 0 || i+1 == stats.Total {
				log.Printf("[%d/%d] ✅ Đã xử lý xong %s", i+1, stats.Total, fileName)
			}
		}
	}

	log.Printf("\n--- KẾT QUẢ IMPORT ---")
	log.Printf("Tổng số file: %d", stats.Total)
	log.Printf("Thành công:   %d", stats.Success)
	log.Printf("Bỏ qua:       %d (Đã có transcript hoặc không tìm thấy video)", stats.Skipped)
	log.Printf("Thất bại:      %d", stats.Failed)

	return nil
}

func parseTSV(filePath string, videoID uuid.UUID) ([]domain.TranscriptSegment, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = '\t'
	reader.LazyQuotes = true

	var segments []domain.TranscriptSegment
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip dòng lỗi định dạng csv
		}

		if len(record) < 3 {
			continue
		}

		start, err1 := strconv.Atoi(record[0])
		end, err2 := strconv.Atoi(record[1])

		// Skip header hoặc dữ liệu không phải số
		if err1 != nil || err2 != nil {
			continue
		}

		segments = append(segments, domain.TranscriptSegment{
			VideoID:     videoID,
			StartTime:   start,
			EndTime:     end,
			TextContent: strings.TrimSpace(record[2]),
		})
	}
	return segments, nil
}
