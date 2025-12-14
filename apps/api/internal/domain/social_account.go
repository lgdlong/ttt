package domain

import (
	"time"

	"github.com/google/uuid"
)

type SocialAccount struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	// 1. Khóa ngoại trỏ về bảng Users (Chủ sở hữu)
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID" json:"-"`

	// 2. Tên nhà cung cấp (google, github, facebook, apple...)
	Provider string `gorm:"type:varchar(50);not null;index:idx_provider_social_id" json:"provider"`

	// 3. ID định danh bên phía Social (QUAN TRỌNG NHẤT)
	// Đây là ID mà Google/Facebook cấp cho user, nó KHÔNG BAO GIỜ đổi.
	// (Lưu ý: Email của user trên Google CÓ THỂ đổi, nhưng ID này thì cố định)
	SocialID string `gorm:"type:varchar(255);not null;index:idx_provider_social_id" json:"social_id"`

	// 4. Email lấy từ Social (để tham khảo, vì đôi khi khác email chính)
	Email string `gorm:"type:varchar(255)" json:"email"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
