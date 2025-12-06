# GitHub Copilot Instructions

Đây là monorepo **TTT** với React frontend và Go backend. Copilot phải tuân thủ các guidelines và skills được định nghĩa trong project.

## Project Overview

- **Frontend:** React + Vite + TypeScript (`apps/web`)
- **Backend:** Go + Gin Framework (`apps/api`)
- **Database:** PostgreSQL
- **Monorepo:** Turborepo + pnpm workspaces

## Skills System

Project này sử dụng custom skills trong `.github/skills/`. Copilot PHẢI tham chiếu và áp dụng các skills phù hợp với context.

### Available Skills

| Skill                | Path                                               | Khi Nào Dùng                                    |
| -------------------- | -------------------------------------------------- | ----------------------------------------------- |
| Backend Development  | `skills/backend-development/`                      | API design, auth, security, database, DevOps    |
| Databases            | `skills/databases/`                                | Schema, queries, migrations, MongoDB/PostgreSQL |
| Frontend Development | `skills/frontend-development/`                     | React components, routing, data fetching        |
| Frontend Design      | `skills/frontend-design/`                          | UI design, styling, animations                  |
| Systematic Debugging | `skills/debugging/systematic-debugging/`           | Bug investigation                               |
| Defense-in-Depth     | `skills/debugging/defense-in-depth/`               | Multi-layer validation                          |
| Root Cause Tracing   | `skills/debugging/root-cause-tracing/`             | Trace errors to source                          |
| Verification         | `skills/debugging/verification-before-completion/` | Verify before claiming done                     |
| Sequential Thinking  | `skills/sequential-thinking/`                      | Complex problem decomposition                   |

## Critical Rules

### 1. Debugging - KHÔNG BAO GIỜ SKIP

Khi gặp bug hoặc test fail:

1. **DỪNG LẠI** - Không đề xuất fix ngay
2. **Đọc error message** cẩn thận
3. **Reproduce** vấn đề
4. **Trace root cause** - Tìm nguồn gốc thực sự
5. **CHỈ SAU ĐÓ** mới đề xuất fix
6. **VERIFY** sau khi fix

> Tham khảo: `skills/debugging/systematic-debugging/SKILL.md`

### 2. Verification Before Completion - BẮT BUỘC

**KHÔNG ĐƯỢC claim "done", "fixed", "works" mà không có evidence từ verification command mới nhất.**

❌ "Should work now"
❌ "Tests should pass"
❌ "Fixed!"

✅ "Ran `pnpm build` - exit code 0"

> Tham khảo: `skills/debugging/verification-before-completion/SKILL.md`

### 3. Frontend Development

Khi tạo React components:

- Sử dụng `React.FC<Props>` pattern
- Lazy load heavy components: `React.lazy(() => import())`
- Wrap trong `<Suspense>` với fallback
- Sử dụng `useSuspenseQuery` cho data fetching
- Follow import aliases: `@/`, `~types`, `~components`, `~features`

> Tham khảo: `skills/frontend-development/SKILL.md`

### 4. Frontend Design

Khi thiết kế UI:

- **KHÔNG** dùng purple gradients on white (AI slop)
- **PHẢI** commit to một aesthetic direction rõ ràng
- **PHẢI** có personality và memorable elements

> Tham khảo: `skills/frontend-design/SKILL.md`

### 5. Backend Development

Khi làm việc với Go backend:

- Follow Gin framework conventions
- Implement proper error handling
- Use structured logging
- Apply OWASP security guidelines
- Write tests cho mọi business logic

> Tham khảo: `skills/backend-development/SKILL.md`

### 6. Database Operations

- Design schema với proper indexing strategy
- Write migrations (up/down)
- Use parameterized queries (KHÔNG raw SQL concatenation)
- Implement connection pooling

> Tham khảo: `skills/databases/SKILL.md`

## Project Structure

```
ttt/
├── apps/
│   ├── api/          # Go Gin backend
│   │   ├── cmd/api/  # Entry point
│   │   └── internal/ # Business logic
│   ├── web/          # React frontend
│   │   └── src/
│   │       ├── components/
│   │       ├── features/
│   │       └── types/
│   └── db/           # Database scripts
├── packages/         # Shared packages
├── .github/
│   └── skills/       # Copilot skills definitions
├── docker-compose.yml
└── turbo.json
```

## Commands Reference

```bash
# Development
pnpm dev              # Run all apps
pnpm dev:web          # Frontend only (port 3000)
pnpm dev:api          # Backend only (port 8080)

# Build
pnpm build            # Build all (React + Go)

# Quality
pnpm lint             # Lint all
pnpm test             # Run tests
pnpm format           # Format code

# Docker
docker compose up -d  # Start all services
```

## Response Guidelines

1. **Luôn tham chiếu skill phù hợp** trước khi trả lời
2. **Code phải production-ready** - không placeholder, không TODO
3. **Giải thích lý do** cho technical decisions
4. **Đề xuất tests** khi implement features
5. **Cảnh báo security issues** khi phát hiện
6. **Verify changes** trước khi claim completion

## Language

- Code: English
- Comments: English hoặc Vietnamese tùy context
- Responses: Vietnamese nếu user hỏi bằng Vietnamese
