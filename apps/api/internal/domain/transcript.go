package domain

import (
	"github.com/google/uuid"
)

// TranscriptSegment đại diện cho một câu thoại trong bảng 'transcript_segments'
type TranscriptSegment struct {
	// Dùng ID số tự tăng (BigInt) vì số lượng record sẽ rất lớn (hàng triệu dòng)
	// UUID ở đây sẽ làm chậm index, nên dùng uint hoặc int64 là tốt nhất
	ID uint `gorm:"primaryKey"`

	// Khóa ngoại trỏ về bảng videos
	VideoID uuid.UUID `gorm:"type:uuid;index;not null"`

	// Thời gian bắt đầu và kết thúc (lưu bằng Mili-giây để chính xác nhất)
	// Index StartTime để query sort theo thời gian hiển thị cực nhanh
	StartTime int `gorm:"index;not null"`
	EndTime   int `gorm:"not null"`

	// Nội dung text hiển thị
	TextContent string `gorm:"type:text;not null"`

	// --- CẤU HÌNH FULL TEXT SEARCH (POSTGRES) ---
	// Field này dùng để mapping với cột tsvector trong Postgres.
	// GORM mặc định sẽ bỏ qua khi insert/update (<-:false) vì ta dùng Trigger DB để tự điền.
	// Nó chỉ dùng để query hoặc migration.
	TSV string `gorm:"column:tsv;type:tsvector;<-:false"`
}

// TableName giúp GORM map đúng vào bảng 'transcript_segments'
func (TranscriptSegment) TableName() string {
	return "transcript_segments"
}

