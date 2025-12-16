# Security Audit Report: TTT Project

**Date:** 2025-12-16
**Auditor:** Senior Security Engineer & Penetration Tester (AI Assistant)

## Summary of Findings

This report details security vulnerabilities discovered during a code audit of the TTT project's Go backend. The audit focused on "low-hanging fruit" and OWASP Top 10 vulnerabilities. Several critical and high-severity issues were identified that require immediate attention.

---

## 🔴 [MỨC ĐỘ: NGHIÊM TRỌNG] - Missing Authentication on Critical Endpoint

- **📍 Vị trí:** `apps/api/internal/routes/routes.go`
- **🐛 Lỗi bảo mật:** **A01:2021 – Broken Access Control**. The endpoint for updating transcript segments is completely public.
- **💡 Giải thích:** The `PATCH /api/v1/transcript-segments/:id` endpoint lacks any authentication or authorization middleware. This allows any unauthenticated user on the internet to modify the text of any transcript segment in the database, leading to data corruption and defacement.
- **🛠️ Cách sửa (Fix):** The route group for this endpoint must be protected with authentication and role-based access control, likely restricting it to moderators and administrators.

```go
// File: apps/api/internal/routes/routes.go

// ... (inside RegisterRoutes function)
		// Transcript segment endpoints (MODERATOR/ADMIN ONLY)
		segments := v1.Group("/transcript-segments")
		segments.Use(middleware.AuthMiddleware(userRepo))
		segments.Use(middleware.RequireMod()) // Only mods or admins can edit
		{
			segments.PATCH("/:id", videoHandler.UpdateSegment)
		}
// ...
```

---

## 🔴 [MỨC ĐỘ: CAO] - Insecure Default JWT Secret

- **📍 Vị trí:** `apps/api/internal/service/auth_service.go`
- **🐛 Lỗi bảo mật:** **A05:2021 – Security Misconfiguration** & **A02:2021 - Cryptographic Failures**.
- **💡 Giải thích:** The `authService` falls back to a hardcoded, well-known, and insecure JWT secret (`"dev-insecure-secret"`) if the `JWT_SECRET` environment variable is not set. While there is a `panic` for production, it relies on `GO_ENV` being set correctly. An attacker who discovers this fallback can forge valid JWT tokens for any user if the application is ever misconfigured in a production-like environment without the `JWT_SECRET` variable.
- **🛠️ Cách sửa (Fix):** Remove the fallback logic entirely. The application should refuse to start if `JWT_SECRET` is not provided, regardless of the environment.

```go
// File: apps/api/internal/service/auth_service.go

func NewAuthService(
	userRepo domain.UserRepository,
	socialAccountRepo domain.SocialAccountRepository,
	sessionRepo domain.SessionRepository,
) domain.AuthService {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// The application should NEVER run with a default secret.
		// Fail fast if the secret is not configured.
		panic("FATAL: JWT_SECRET environment variable is not set.")
	}

	// ... (rest of the function)
}
```

---

## 🔴 [MỨC ĐỘ: TRUNG BÌNH] - Root User in Development Docker Container

- **📍 Vị trí:** `apps/api/Dockerfile.dev`
- **🐛 Lỗi bảo mật:** **Principle of Least Privilege**.
- **💡 Giải thích:** The development Docker container (`Dockerfile.dev`) runs its process as the `root` user. While intended for development, this is a bad practice. An exploit in the application or one of its dependencies during development could grant an attacker root access within the container, allowing them to install malicious software or tamper with the development environment.
- **🛠️ Cách sửa (Fix):** Add a non-root user to the `Dockerfile.dev` and switch to it before running the application, mirroring the best practice already implemented in the production `Dockerfile`.

```dockerfile
# File: apps/api/Dockerfile.dev

# =============================================================================
# Development Dockerfile for API (Hot Reload with Air)
# =============================================================================
FROM golang:1.24-alpine

# Install development dependencies
RUN apk add --no-cache git

WORKDIR /app

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy go mod files
COPY --chown=appuser:appgroup go.mod go.sum ./
RUN go mod download

# Copy source code
COPY --chown=appuser:appgroup . .

# Switch to the non-root user
USER appuser

# Expose API port
EXPOSE 8080

# Start with Air for hot reload
CMD ["go", "run", "github.com/air-verse/air@latest", "-c", ".air.toml"]
```

---

## ✅ Checks Passed

- **SQL Injection:** No clear instances of SQL injection were found. The code consistently uses parameterized queries (`?` or `$1`) or GORM's safe methods, which correctly separate SQL code from user-provided data.
- **Frontend XSS:** A review of the frontend code is required to check for XSS vulnerabilities, particularly around the use of `dangerouslySetInnerHTML`. This was not performed as the frontend files were not in scope for this audit.
- **CORS Configuration:** The CORS middleware in `apps/api/internal/middleware/cors.go` dynamically reflects the `Origin` header if it matches the `ALLOWED_ORIGIN` environment variable. This is safe as long as `ALLOWED_ORIGIN` is configured to a specific, trusted domain in production and not a wildcard.
