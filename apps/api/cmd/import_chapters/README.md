# Import Chapters & Summary Tool

Công cụ CLI này được sử dụng để tự động nhập dữ liệu tóm tắt (summary) và các chương (chapters) của video từ các tệp JSON vào cơ sở dữ liệu PostgreSQL.

## Chức năng chính

- Đọc các tệp JSON từ một thư mục chỉ định.
- Trích xuất YouTube ID từ tên tệp (11 ký tự đầu tiên).
- Cập nhật trường `summary` cho video tương ứng trong bảng `videos`.
- Xóa các chương cũ và chèn các chương mới vào bảng `video_chapters`.
- Sử dụng Database Transaction để đảm bảo tính toàn vẹn dữ liệu.

## Cách sử dụng

Chạy lệnh sau từ thư mục `apps/api`:

```bash
go run cmd/import_chapters/main.go -dir <đường_dẫn_thư_mục_chứa_json>
```

**Tham số:**
- `-dir`: (Tùy chọn) Đường dẫn đến thư mục chứa các tệp JSON. Mặc định là `../../agents/resources/transcript_to_json/output`.

## Quy ước đặt tên tệp

Tên tệp JSON phải bắt đầu bằng 11 ký tự của YouTube ID.
Ví dụ: `_2oihXPeUwQ_Title.json` -> YouTube ID sẽ là `_2oihXPeUwQ`.

## Định dạng tệp JSON đầu vào

Công cụ yêu cầu cấu trúc JSON như sau:

```json
{
  "analysis": {
    "summary": "Nội dung tóm tắt của video..."
  },
  "transcript": [
    {
      "segment_id": 1,
      "title": "Tiêu đề chương 1",
      "content": "Nội dung chi tiết của chương 1",
      "start_time": 0
    },
    {
      "segment_id": 2,
      "title": "Tiêu đề chương 2",
      "content": "Nội dung chi tiết của chương 2",
      "start_time": 120
    }
  ]
}
```

## Lưu ý

- Video phải tồn tại trong cơ sở dữ liệu trước khi thực hiện import. Nếu không tìm thấy YouTube ID, công cụ sẽ bỏ qua tệp đó và log cảnh báo.
- Mọi chương cũ của video trong DB sẽ bị xóa và thay thế bằng dữ liệu mới từ tệp JSON.
