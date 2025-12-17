## TỔNG QUAN
Dự án có nền tảng Backend rất mạnh về **Search** và **Tagging (đa ngôn ngữ)** nhờ kiến trúc `Canonical-Alias` và tích hợp `pgvector`. Tuy nhiên, các tính năng liên quan đến **AI Automation** (tự động hóa quy trình) và **User Interaction** (tương tác người dùng cuối) hầu như chưa được triển khai hoặc chỉ dừng lại ở mức cơ sở hạ tầng.

## CHI TIẾT ĐÁNH GIÁ

| Tính năng | Trạng thái | Minh chứng (File/Hàm tìm thấy) | Thành phần còn thiếu/Cần làm |
| :--- | :--- | :--- | :--- |
| **Tìm kiếm theo Title/Tags** | **Hoàn thành** | `video_repository.go`: Hàm `GetVideoList` hỗ trợ filter `query` (title) và `tag_id`. `GlobalSearchBar.tsx` ở Frontend đã tích hợp. | Đã đầy đủ chức năng cơ bản. |
| **Tags/Category Song ngữ** | **Một phần** | `tag_service_v2.go`, `004_create_canonical_alias_tags.sql`: Hệ thống `CanonicalTag` (Concept) + `TagAlias` (Keywords) hỗ trợ đa ngôn ngữ cực tốt. | **Thiếu Category:** Hệ thống chỉ dùng Tags phẳng, không có bảng/cấu trúc cho Category phân cấp. |
| **AI Tự động tạo Tag** | **Chưa có** | `tag_service_v2.go`: Có logic `ResolveTag` dùng vector search để tìm tag, NHƯNG `CreateVideo` (`video_handler.go`) chỉ lưu metadata YouTube thô sơ. | Chưa có "AI Worker" hoặc hook để tự động gọi `ResolveTag` dựa trên Title khi video vừa được thêm vào. |
| **AI Tóm tắt nội dung** | **Chưa có** | Đã tìm trong `openai.go` và `video_service.go`. Không có hàm nào liên quan đến `Summarize`. | Cần thêm prompt tóm tắt vào `openai.go` và API endpoint để gọi nó. |
| **AI Gắn Tag từ Script** | **Chưa có** | `video_repository.go`: Có `SearchTranscripts` (Full-text search) nhưng không có logic phân tích ngược lại để sinh ra Tags. | Cần logic đọc toàn bộ transcript -> gửi cho AI -> trích xuất keywords -> map vào Tag Canonical. |
| **Quản lý Version Script** | **Chưa có** | DB chỉ có `transcript_segments` (dữ liệu hiện tại) và `video_transcript_reviews` (log người duyệt). Không có bảng `versions`. | Cần tạo bảng `transcript_versions` để lưu snapshot mỗi khi API `UpdateSegment` được gọi. |
| **So sánh Diff Script** | **Chưa có** | Rà soát toàn bộ code không thấy từ khóa `diff` hay thư viện so sánh text nào. | Cần thư viện diff ở Frontend hoặc Backend để so sánh 2 phiên bản text. |
| **Hệ thống duyệt Script** | **Hoàn thành** | `video_transcript_reviews` table, `VideoTranscriptReviewHandler`, `GetReviewCountsForVideos`. | Logic đã hoạt động: Mod submit review -> tăng count -> hiển thị badge "Đã duyệt". |
| **Ghi chú kèm Timestamp** | **Chưa có** | `apps/api/internal/dto/video.go`: Chỉ có field `Notes` trong `SubmitReviewRequest` (cho Mod), không phải cho User. | Cần bảng `user_video_notes` (video_id, user_id, timestamp, content) và API tương ứng. |

## KẾT LUẬN & ĐỀ XUẤT

**1. Điểm mạnh:**
*   Hệ thống Tagging V2 (Canonical/Alias) được thiết kế rất thông minh, sẵn sàng cho việc mở rộng đa ngôn ngữ và tìm kiếm ngữ nghĩa (Semantic Search).
*   Search Engine (Hybrid: SQL LIKE + Vector Search) đã được triển khai tốt.

**2. Điểm yếu (Cần ưu tiên xử lý):**
*   **Thiếu Automation:** Mặc dù "có súng" (OpenAI client, Vector DB) nhưng "chưa bóp cò". Việc thêm video hiện tại hoàn toàn thủ công, chưa tận dụng AI để auto-tagging.
*   **Thiếu Safety cho Script:** Việc sửa Script ghi đè trực tiếp vào DB mà không có backup (Versioning) là rất rủi ro cho một hệ thống crowdsourcing/mod.

**3. Roadmap đề xuất:**
1.  **Script Versioning:** Tạo bảng lưu lịch sử thay đổi ngay lập tức.
2.  **Auto-Tagging Pipeline:** Viết một function chạy background: Khi `CreateVideo` thành công -> Gọi OpenAI phân tích Title -> Gọi `ResolveTag` để tự động gắn tag.
3.  **Category:** Quyết định xem có cần bảng Category riêng không, hay quy ước một số Tag đặc biệt làm Category.