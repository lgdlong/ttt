## 1. Xác thực và Phân quyền (Authentication & Authorization)

Đây là nơi hacker thường "ghé thăm" đầu tiên.

* **JWT Security:** Nếu dùng JWT, hãy đảm bảo:
* **Secret Key** đủ mạnh và không bị hardcode trong code.
* Có cơ chế **Rotate Refresh Token** và thu hồi token (Blacklist) khi cần.
* Sử dụng flag `HttpOnly` và `Secure` cho Cookie nếu lưu token ở đó để chống XSS.


* **Broken Access Control:** Kiểm tra xem một User A có thể sửa/xóa bài viết của User B bằng cách thay đổi `ID` trên URL hay API request không (lỗi IDOR).
* **Rate Limiting:** Chặn các cuộc tấn công Brute-force vào endpoint `/login` hoặc `/register`.

## 2. Kiểm soát dữ liệu đầu vào (Input Validation & Sanitization)

"Đừng bao giờ tin tưởng người dùng" là nguyên tắc vàng.

* **SQL Injection:** Bạn đang dùng Go, hãy chắc chắn sử dụng **parameterized queries** (truy vấn có tham số). Tuyệt đối không cộng chuỗi để tạo SQL query.
* **XSS (Cross-Site Scripting):** React mặc định đã chống XSS khá tốt, nhưng hãy cẩn thận với `dangerouslySetInnerHTML`. Mọi dữ liệu từ User hiển thị lên màn hình cần được sanitize.
* **Validation:** Sử dụng các thư viện như `go-playground/validator` ở Backend để đảm bảo dữ liệu gửi lên đúng định dạng (email, độ dài pass, type...).

## 3. Bảo mật giao thức và Header (Security Headers)

Đây là lớp "giáp" ngăn chặn nhiều kiểu tấn công trình duyệt.

* **CORS (Cross-Origin Resource Sharing):** Chỉ cho phép các Domain cụ thể (ví dụ: `yourdomain.com`) gọi API, đừng để `Allow-Origin: *`.
* **Security Headers:** Cấu hình các header quan trọng:
* `Content-Security-Policy (CSP)`: Ngăn chặn load script lạ.
* `Strict-Transport-Security (HSTS)`: Ép trình duyệt luôn dùng HTTPS.
* `X-Content-Type-Options: nosniff`.


* **HTTPS:** Đảm bảo toàn bộ traffic được mã hóa qua TLS/SSL.

---

## 4. Quản lý bí mật (Secrets Management)

* **Environment Variables:** Tuyệt đối không commit file `.env` lên GitHub.
* **Git History:** Kiểm tra xem trong lịch sử commit cũ có lỡ để lộ DB Password hay API Key nào không. Nếu có, hãy dùng `git-filter-repo` để xóa hoặc đổi key mới ngay lập tức.

## 5. Logging và Monitoring

* **Structured Logging:** Log lại các hành vi đáng ngờ (ví dụ: 1 IP login sai 50 lần).
* **Error Handling:** Đừng trả về nguyên văn lỗi của Database (như `Table 'users' not found`) cho Client. Hacker sẽ dựa vào đó để biết cấu trúc DB của bạn. Hãy trả về mã lỗi chung chung như `Internal Server Error`.

# ĐỊNH DẠNG TRẢ LỜI (RESPONSE FORMAT)

Nếu phát hiện lỗi, hãy trình bày theo format sau:

1.  **🔴 [MỨC ĐỘ: NGHIÊM TRỌNG/CAO/TRUNG BÌNH]**
2.  **📍 Vị trí:** `Tên file : Số dòng`
3.  **🐛 Lỗi bảo mật:** Tên lỗi (ví dụ: SQL Injection).
4.  **💡 Giải thích:** Tại sao đoạn này nguy hiểm (ngắn gọn).
5.  **🛠️ Cách sửa (Fix):** Cung cấp đoạn code đã sửa hoàn chỉnh (Production-ready).

---
**LƯU Ý:**
- Nếu code an toàn, hãy nói ngắn gọn: "✅ Không phát hiện lỗ hổng bảo mật rõ ràng trong đoạn này."
- Không đưa ra các lời khuyên chung chung (như "nên viết clean code") trừ khi nó ảnh hưởng trực tiếp đến bảo mật.