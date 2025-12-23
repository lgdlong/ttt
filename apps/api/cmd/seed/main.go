package main

import (
	"api/cmd/seed/function"
	"api/internal/database"
	"flag"
	"log"
	"os"
)

func main() {
	// Định nghĩa các flags để chọn chạy cái gì
	jsonFile := flag.String("json", "../../tsv_files/ttt_without_drugs_videos.json", "Đường dẫn file JSON metadata video")
	tsvDir := flag.String("tsv", "../../tsv_files/ttt-3", "Thư mục chứa file TSV")
	action := flag.String("action", "all", "Chọn action: videos, transcripts, all")
	force := flag.Bool("force", false, "Ghi đè transcript nếu đã tồn tại")
	flag.Parse()

	// 1. Kết nối DB
	// Lưu ý: Đảm bảo biến môi trường DB_HOST=localhost nếu chạy từ máy ngoài Docker
	log.Println("🔌 Đang kết nối Database...")
	dbService := database.New()
	if dbService == nil {
		log.Fatal("Không thể khởi tạo database service")
	}
	gormDB := dbService.GetGormDB()

	var err error

	// 2. Chạy Import Videos
	if *action == "videos" || *action == "all" {
		if _, err := os.Stat(*jsonFile); os.IsNotExist(err) {
			log.Fatalf("File JSON không tồn tại: %s", *jsonFile)
		}
		err = function.ImportVideos(gormDB, *jsonFile)
		if err != nil {
			log.Fatalf("Lỗi Import Videos: %v", err)
		}
	}

	// 3. Chạy Import Transcripts
	if *action == "transcripts" || *action == "all" {
		if _, err := os.Stat(*tsvDir); os.IsNotExist(err) {
			log.Fatalf("Thư mục TSV không tồn tại: %s", *tsvDir)
		}
		err = function.ImportTranscripts(gormDB, *tsvDir, *force)
		if err != nil {
			log.Fatalf("Lỗi Import Transcripts: %v", err)
		}
	}
}
