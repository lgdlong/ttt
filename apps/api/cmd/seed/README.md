# Video & Transcript Seeder

Công cụ CLI này dùng để nạp dữ liệu (seed) từ các file JSON và TSV vào Database. Nó hỗ trợ nạp thông tin metadata của video và nội dung transcript (phân đoạn thoại).

## Tính năng mới & Cải tiến
- **Thông minh (Skipping)**: Tự động bỏ qua các file TSV của video đã có transcript trong database (dựa trên field `has_transcript`), giúp tiết kiệm thời gian khi resume hoặc chạy lại.
- **An toàn (Atomic per-file)**: Mỗi file TSV được xử lý trong một Database Transaction riêng. Nếu một file bị lỗi, dữ liệu của video đó sẽ không bị lưu dở dang, đảm bảo tính toàn vẹn.
- **Ép buộc (Force Update)**: Hỗ trợ xóa và nạp lại transcript thông qua flag `-force`.
- **Báo cáo chi tiết**: Kết thúc quá trình sẽ có bảng thống kê số lượng file thành công, thất bại và số file được bỏ qua.

## Yêu cầu chuẩn bị

### 1. Database
Đảm bảo bạn đã cấu hình biến môi trường chính xác (thông thường là file `.env` tại `apps/api`). Nếu chạy script từ máy local (không phải trong container), hãy đảm bảo `DB_HOST=localhost`.

### 2. Định dạng dữ liệu

#### File JSON Video (`clean_videos.json`):
```json
[
  {
    "youtube_id": "dQw4w9WgXcQ",
    "title": "Never Gonna Give You Up",
    "published_at": "2009-10-25",
    "duration": 212,
    "view_count": 1000000,
    "thumbnail_url": "https://..."
  }
]
```

#### File TSV Transcript:
Tên file phải trùng với `youtube_id` (ví dụ: `dQw4w9WgXcQ.tsv`). Định dạng phân cách bằng tab (`\t`): 
```tsv
start	end	text
0	5000	Chào mừng các bạn đã quay trở lại
5000	10000	Hôm nay chúng ta sẽ nói về...
```
*Lưu ý: Thời gian tính bằng mili-giây.*

## Cách sử dụng

Di chuyển vào thư mục `apps/api` và chạy lệnh:

### 1. Nạp tất cả (Videos + Transcripts)
```bash
go run cmd/seed/*.go -action=all
```

### 2. Chỉ nạp danh sách Video
```bash
go run cmd/seed/*.go -action=videos -json="../../tsv_files/my_videos.json"
```

### 3. Chỉ nạp Transcript (Resume)
Mặc định script sẽ bỏ qua những video đã có transcript.
```bash
go run cmd/seed/*.go -action=transcripts -tsv="../../tsv_files/my_tsv_folder"
```

### 4. Nạp lại Transcript (Overwrite)
Nếu bạn muốn xóa transcript cũ và nạp lại từ file mới:
```bash
go run cmd/seed/*.go -action=transcripts -force=true
```

## Các Flags hỗ trợ
- `-action`: Hành động thực hiện (`videos`, `transcripts`, `all`). Mặc định là `all`.
- `-json`: Đường dẫn đến file JSON metadata video.
- `-tsv`: Đường dẫn đến thư mục chứa các file `.tsv`.
- `-force`: (`true`/`false`) Nếu là true, sẽ xóa dữ liệu cũ và nạp lại nếu video đã có transcript. Mặc định là `false`.

## Lưu ý về hiệu năng
- Công cụ sử dụng `Batch Insert` (1000 dòng/lần) bên trong transaction của từng file để tối ưu tốc độ.
- Hệ thống sẽ log tiến độ sau mỗi 50 file được xử lý xong.