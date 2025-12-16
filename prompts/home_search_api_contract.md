Dưới đây là bản **API Contract (Thiết kế giao diện lập trình)** chi tiết dạng plain text. Bạn có thể copy toàn bộ nội dung này và gửi cho AI Agent hoặc Developer của bạn để họ triển khai chính xác logic tìm kiếm hợp nhất (Unified Search) cho trang chủ.

-----

# 📋 API DESIGN CONTRACT: UNIFIED HOMEPAGE SEARCH

**Feature:** Tìm kiếm video theo Tiêu đề (Title) HOẶC Tags (Display Name) thông qua một thanh search duy nhất.

## 1\. Endpoint Specification

  * **Method:** `GET`
  * **URL:** `/api/v1/videos`
  * **Access:** Public (Không yêu cầu Authentication)

## 2\. Request Parameters (Query Params)

| Param | Type | Required | Default | Mô tả chi tiết (Logic xử lý) |
| :--- | :--- | :--- | :--- | :--- |
| **`q`** | `string` | No | `null` | **(Cập nhật mới)** Từ khóa tìm kiếm tự do.<br>Logic: Tìm các video mà `q` xuất hiện trong **Title** HOẶC **Tag Name**.<br>Ví dụ: `q=java` sẽ trả về video có title "Học Java" VÀ video có tag "Java Core". |
| `page` | `int` | No | `1` | Số trang hiện tại. |
| `limit` | `int` | No | `10` | Số lượng video/trang. |
| `sort` | `string` | No | `newest` | `newest`, `popular` (view), `views`. |
| `tags` | `string`| No | `null` | (Giữ nguyên logic cũ) Lọc theo danh sách Tag ID cụ thể (comma-separated). Nếu dùng kết hợp với `q`, logic là AND (Tìm `q` trong tập video đã lọc theo `tags`). |
| `has_transcript`| `bool` | No | `null` | Lọc video có/không có phụ đề. |

## 3\. Backend Implementation Logic (Yêu cầu cho Dev/AI)

Agent cần cập nhật tầng **Repository** (`internal/repository/video_repository.go`) theo luồng dữ liệu sau:

### 3.1. Query Construction

Khi tham số `q` được gửi lên (không rỗng):

1.  **JOIN Tables:** Thực hiện `LEFT JOIN` từ bảng `videos` sang bảng `video_canonical_tags`, và từ đó JOIN sang `canonical_tags`.
2.  **Filter Condition (WHERE Clause):**
      * Sử dụng nhóm điều kiện `OR`.
      * Pseudocode SQL: `WHERE (LOWER(videos.title) LIKE %q% OR LOWER(canonical_tags.display_name) LIKE %q%)`.
3.  **Deduplication (Quan trọng):**
      * Bắt buộc sử dụng `GROUP BY videos.id`.
      * **Lý do:** Một video có thể khớp cả Title lẫn nhiều Tag. Nếu không Group, kết quả trả về sẽ bị duplicate video đó nhiều lần.

### 3.2. Performance Note

  * Sử dụng `ILIKE` (PostgreSQL) hoặc `LOWER()` để tìm kiếm không phân biệt hoa thường.
  * Nên Preload `CanonicalTags` để Frontend hiển thị được danh sách tag ngay trên card video (giúp user hiểu tại sao video này hiện ra dù title không chứa từ khóa).

## 4\. Response Format (JSON)

Status: `200 OK`

```json
{
  "success": true,
  "message": "Videos retrieved successfully",
  "data": [
    {
      "id": "uuid-video-1",
      "title": "Hướng dẫn lập trình Golang cơ bản",
      "thumbnail_url": "https://img.youtube.com/...",
      "duration": 600,
      "view_count": 1500,
      "published_at": "2023-12-01",
      "has_transcript": true,
      // Frontend hiển thị list tags này.
      // VD: Nếu search "Backend", video này hiện ra nhờ tag "Backend" bên dưới
      "tags": [
        {
          "id": "uuid-tag-1",
          "name": "Backend"
        },
        {
          "id": "uuid-tag-2",
          "name": "Golang"
        }
      ]
    },
    {
      "id": "uuid-video-2",
      "title": "Backend Roadmap 2024 (Khớp do Title)",
      "tags": []
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total_items": 45,
    "total_pages": 5
  }
}
```

## 5\. Test Cases (Tiêu chí chấp nhận)

1.  **Case: Search Title**
      * Input: `q=roadmap`
      * Expect: Trả về video có chữ "Roadmap" trong tiêu đề.
2.  **Case: Search Tag**
      * Input: `q=money`
      * Expect: Trả về video có tiêu đề tiếng Việt "Cách làm giàu" NHƯNG được gắn tag "Money".
3.  **Case: Search Combined**
      * Input: `q=java`
      * Expect: Trả về cả video title "Học Java" và video title "Spring Boot" (có tag Java).
4.  **Case: Duplicate Check**
      * Input: `q=test` (Video vừa có title "Test", vừa có tag "Test")
      * Expect: Video đó chỉ xuất hiện **1 lần duy nhất** trong danh sách `data`.