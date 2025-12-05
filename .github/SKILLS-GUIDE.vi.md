# Hướng Dẫn Sử Dụng GitHub Copilot Skills

## Giới Thiệu

Skills là hệ thống mở rộng kiến thức cho GitHub Copilot Agent, giúp AI hiểu sâu hơn về project và cung cấp responses chất lượng cao hơn. Skills được định nghĩa trong `.github/skills/` và tự động load khi Copilot làm việc với codebase.

## Cấu Trúc Skills

```
.github/skills/
├── backend-development/     # Phát triển backend (Go, Node.js, Python)
│   ├── SKILL.md
│   └── references/          # Tài liệu tham khảo chi tiết
├── databases/               # MongoDB & PostgreSQL
│   ├── SKILL.md
│   ├── references/
│   └── scripts/             # Scripts backup, migrate, performance
├── debugging/               # Kỹ thuật debug chuyên sâu
│   ├── defense-in-depth/    # Validation nhiều lớp
│   ├── root-cause-tracing/  # Truy vết nguyên nhân gốc
│   ├── systematic-debugging/# Debug có hệ thống
│   └── verification-before-completion/
├── frontend-design/         # Thiết kế UI đẹp, độc đáo
│   └── SKILL.md
├── frontend-development/    # React + TypeScript patterns
│   ├── SKILL.md
│   └── resources/
└── sequential-thinking/     # Tư duy tuần tự cho vấn đề phức tạp
    ├── SKILL.md
    └── references/
```

## Các Skills Có Sẵn

### 1. Backend Development
**Khi nào dùng:** Thiết kế API, authentication, database optimization, security, DevOps

**Trigger keywords:** "tạo API", "authentication", "optimize query", "security", "Docker", "Kubernetes"

**Nội dung:**
- Lựa chọn technology stack (Go, Node.js, Python, Rust)
- REST/GraphQL/gRPC API design
- OAuth 2.1 + JWT authentication
- OWASP Top 10 security
- Database optimization
- CI/CD pipelines

### 2. Databases
**Khi nào dùng:** Schema design, queries, indexing, migrations, backup/restore

**Trigger keywords:** "database", "MongoDB", "PostgreSQL", "query", "migration", "backup"

**Nội dung:**
- MongoDB: CRUD, Aggregation, Atlas, Indexing
- PostgreSQL: Queries, Administration, Performance
- Scripts tự động: backup, migrate, performance check

### 3. Frontend Development
**Khi nào dùng:** Tạo components, pages, features trong React/TypeScript

**Trigger keywords:** "component", "React", "TypeScript", "styling", "routing"

**Nội dung:**
- React patterns với Suspense, lazy loading
- TanStack Query (data fetching)
- TanStack Router
- MUI v7 styling
- File organization theo features
- TypeScript best practices

### 4. Frontend Design
**Khi nào dùng:** Cần UI đẹp, độc đáo, không generic

**Trigger keywords:** "design", "UI đẹp", "giao diện", "animation", "styling"

**Nội dung:**
- Typography độc đáo (không dùng Inter, Arial)
- Color palette có cá tính
- Motion & animations (anime.js)
- Layout sáng tạo
- Tránh "AI slop" aesthetics

### 5. Debugging Skills

#### a) Systematic Debugging
**Khi nào dùng:** Gặp bug, test fail, behavior lạ

**Quy tắc vàng:** KHÔNG SỬA KHI CHƯA TÌM RA NGUYÊN NHÂN GỐC

**4 giai đoạn:**
1. Root Cause Investigation (đọc error, reproduce, check changes)
2. Hypothesis Formation
3. Fix Implementation
4. Verification

#### b) Defense-in-Depth
**Khi nào dùng:** Data không hợp lệ gây lỗi sâu trong execution

**Nguyên tắc:** Validate ở MỌI layer data đi qua

**4 layers:**
1. Entry Point Validation
2. Business Logic Validation
3. Environment Guards
4. Debug Instrumentation

#### c) Root Cause Tracing
**Khi nào dùng:** Error xảy ra sâu trong call stack

**Nguyên tắc:** Trace NGƯỢC về trigger gốc, không fix ở symptom

#### d) Verification Before Completion
**Khi nào dùng:** Trước khi claim "done", commit, hoặc tạo PR

**Quy tắc:** KHÔNG CLAIM MÀ KHÔNG CÓ EVIDENCE TỪ VERIFICATION MỚI NHẤT

### 6. Sequential Thinking
**Khi nào dùng:** Vấn đề phức tạp cần suy luận từng bước

**Trigger keywords:** "phân tích", "lên kế hoạch", "thiết kế kiến trúc", "decompose"

**Capabilities:**
- Iterative reasoning
- Dynamic scope adjustment
- Revision tracking
- Branch exploration

## Cách Sử Dụng

### Tự động (Recommended)
Skills tự động được áp dụng khi Copilot detect context phù hợp qua `copilot-instructions.md`.

### Explicit Request
Nếu muốn áp dụng skill cụ thể:

```
@workspace Áp dụng skill backend-development để thiết kế API cho user authentication
```

```
@workspace Dùng systematic-debugging để debug test fail này
```

```
@workspace Thiết kế UI cho landing page theo frontend-design guidelines
```

### Reference Files
Mỗi skill có thư mục `references/` chứa tài liệu chi tiết. Copilot tự động reference khi cần.

## Best Practices

### 1. Debugging
- **Luôn** tìm root cause trước khi fix
- **Luôn** verify sau khi fix
- **Không** claim done mà không có evidence

### 2. Development
- Follow file organization patterns
- Sử dụng import aliases (@/, ~types, ~components)
- Lazy load heavy components

### 3. Design
- Tránh generic fonts (Inter, Arial, Roboto)
- Commit to một aesthetic direction rõ ràng
- Dùng animations có chủ đích

### 4. Sequential Thinking
- Bắt đầu với estimate rough, refine dần
- Sử dụng revision khi assumptions sai
- Branch khi có nhiều approaches

## Cập Nhật Skills

1. Tạo/edit file trong `.github/skills/`
2. Mỗi skill cần `SKILL.md` với frontmatter:
   ```yaml
   ---
   name: skill-name
   description: Mô tả ngắn gọn
   when_to_use: Khi nào áp dụng
   ---
   ```
3. Thêm `references/` cho tài liệu chi tiết
4. Commit và push

## Troubleshooting

**Skill không được áp dụng:**
- Kiểm tra `copilot-instructions.md` có reference đúng path
- Đảm bảo frontmatter YAML hợp lệ
- Thử explicit request

**Copilot không theo guidelines:**
- Nhắc lại skill cần áp dụng
- Quote specific rules từ SKILL.md
- Chia nhỏ request

---

*Skills system version: 1.0.0*
*Last updated: December 2025*
