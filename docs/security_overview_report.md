# Security Overview Report - TTT Project
**Ngày kiểm tra:** 18 tháng 12, 2025  
**Phạm vi:** Backend Go (Gin) + Frontend React + Database PostgreSQL  
**Chuẩn tham chiếu:** OWASP Top 10, security_check_overview.md

---

## 📊 EXECUTIVE SUMMARY

Project TTT đã implement nhiều best practices về bảo mật cơ bản, nhưng vẫn còn **một số lỗ hổng nghiêm trọng** cần được xử lý ngay lập tức trước khi deploy production.

### Điểm mạnh ✅
- JWT authentication được implement đúng cách với HttpOnly cookies
- Sử dụng bcrypt cho password hashing
- Input validation với Gin binding tags
- Parameterized queries ngăn SQL injection
- Refresh token rotation mechanism
- Session management với blacklist
- React không có `dangerouslySetInnerHTML` (chống XSS)

### Lỗ hổng nghiêm trọng 🔴
1. **KHÔNG có Rate Limiting** trên `/auth/login` và `/auth/signup`
2. **KHÔNG có Security Headers** (CSP, HSTS, X-Frame-Options)
3. **Leak thông tin nhạy cảm** qua error messages
4. **Default JWT secret fallback** trong middleware (development mode)

---

## 1. XÁC THỰC VÀ PHÂN QUYỀN (Authentication & Authorization)

### ✅ ĐÃ ĐẢM BẢO

#### 1.1 JWT Security
**File:** [apps/api/internal/service/auth_service.go](apps/api/internal/service/auth_service.go#L43-L47)
```go
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    panic("FATAL: JWT_SECRET is not set")
}
```
- ✅ **FAIL FAST** nếu JWT_SECRET không được set
- ✅ Secret key được load từ environment variable
- ✅ Không hardcode trong code

#### 1.2 JWT Token Cookies - HttpOnly & Secure
**File:** [apps/api/internal/handler/auth_handler.go](apps/api/internal/handler/auth_handler.go#L23-L36)
```go
func (h *AuthHandler) setAuthCookie(c *gin.Context, token string) {
    secure := os.Getenv("ENV") == "production"
    c.SetCookie(
        "token",    
        token,      
        60*60*24*7, // 7 days
        "/",        
        "",         
        secure,     // ✅ HTTPS only in production
        true,       // ✅ HttpOnly: chống XSS
    )
}
```
- ✅ **HttpOnly flag** = true → JavaScript không thể access cookie (chống XSS)
- ✅ **Secure flag** = true trong production → Chỉ gửi qua HTTPS
- ✅ Refresh token cũng được lưu tương tự

#### 1.3 Access Control - Role-based
**File:** [apps/api/internal/middleware/auth.go](apps/api/internal/middleware/auth.go#L143-L176)
```go
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
    // Check if user's role is in allowed roles
    for _, r := range allowedRoles {
        if string(userRole) == r {
            allowed = true
            break
        }
    }
    if !allowed {
        c.JSON(http.StatusForbidden, dto.ErrorResponse{
            Error:   "Forbidden",
            Message: "Insufficient permissions",
            Code:    http.StatusForbidden,
        })
        c.Abort()
        return
    }
}
```
- ✅ Middleware kiểm tra role trước khi cho phép access
- ✅ Có `RequireAdmin()` và `RequireMod()` helpers

#### 1.4 Refresh Token Rotation & Session Blacklist
**File:** [apps/api/internal/service/auth_service.go](apps/api/internal/service/auth_service.go#L210-L250)
```go
func (s *authService) RefreshToken(refreshToken string) (*dto.AuthResponse, error) {
    session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
    if session.IsBlocked {
        return nil, errors.New("session is blocked")
    }
    if session.ExpiresAt.Before(time.Now()) {
        return nil, errors.New("session expired")
    }
    // ...
}
```
- ✅ Kiểm tra session có bị block không
- ✅ Kiểm tra expiration
- ✅ Có chức năng `LogoutAll()` để revoke tất cả session của user

---

### 🔴 LỖ HỔNG NGHIÊM TRỌNG

#### 🔴 [CAO] 1.1: KHÔNG CÓ RATE LIMITING - Dễ bị Brute-force Attack

**📍 Vị trí:**  
- [apps/api/internal/handler/auth_handler.go](apps/api/internal/handler/auth_handler.go#L90): Endpoint `/auth/login`
- [apps/api/internal/handler/auth_handler.go](apps/api/internal/handler/auth_handler.go#L138): Endpoint `/auth/signup`

**🐛 Lỗi bảo mật:**  
Các endpoint authentication KHÔNG có rate limiting. Hacker có thể:
- Brute-force password với hàng ngàn request/giây
- Dictionary attack để đoán password phổ biến
- Account enumeration để tìm username/email hợp lệ
- DDoS endpoint `/auth/login` để làm sập service

**💡 Giải thích:**  
Không có middleware nào kiểm soát số lượng request từ cùng một IP. Ví dụ attacker có thể:
```bash
# Thử 10,000 passwords trong 1 phút
for i in {1..10000}; do
  curl -X POST http://api/auth/login -d '{"username":"admin","password":"pass'$i'"}' &
done
```

**🛠️ Cách sửa (Fix):**

**Bước 1:** Install rate limiting library:
```bash
cd apps/api
go get github.com/ulule/limiter/v3
go get github.com/ulule/limiter/v3/drivers/store/memory
```

**Bước 2:** Tạo middleware rate limiting:  
**File mới:** `apps/api/internal/middleware/rate_limit.go`
```go
package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitAuth creates rate limiter for auth endpoints
// Limit: 5 requests per minute per IP
func RateLimitAuth() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  5, // Max 5 login attempts per minute
	}

	store := memory.NewStore()
	instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))

	middleware := mgin.NewMiddleware(instance, mgin.WithKeyGetter(func(c *gin.Context) string {
		// Rate limit by IP address
		return c.ClientIP()
	}))

	return middleware
}

// RateLimitGeneral creates rate limiter for general API endpoints
// Limit: 100 requests per minute per IP
func RateLimitGeneral() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100,
	}

	store := memory.NewStore()
	instance := limiter.New(store, rate)

	middleware := mgin.NewMiddleware(instance)
	return middleware
}
```

**Bước 3:** Apply vào routes:  
**File:** `apps/api/internal/routes/routes.go`
```go
func RegisterRoutes(/* ... */) {
	// Apply general rate limit to all routes
	router.Use(middleware.RateLimitGeneral())

	// Auth routes với stricter rate limit
	authGroup := router.Group("/api/auth")
	authGroup.Use(middleware.RateLimitAuth()) // ✅ 5 req/min
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/signup", authHandler.Signup)
		// ...
	}
}
```

**Verify:**
```bash
# Test rate limiting
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"wrong"}'
  echo ""
done

# Expected: Sau 5 request, server trả về 429 Too Many Requests
```

---

#### 🟡 [TRUNG BÌNH] 1.2: Default JWT Secret trong Middleware (Development)

**📍 Vị trí:** [apps/api/internal/middleware/auth.go](apps/api/internal/middleware/auth.go#L17-L19)

**🐛 Lỗi bảo mật:**
```go
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    jwtSecret = "default-secret-change-in-production" // ⚠️ XẤU
}
```
Nếu developer quên set `JWT_SECRET` trong .env, app vẫn chạy với default secret dễ đoán.

**💡 Giải thích:**  
Middleware này được dùng để verify JWT token. Nếu secret bị lộ, attacker có thể forge token bất kỳ và truy cập vào bất kỳ tài khoản nào.

**🛠️ Cách sửa:**
```go
func AuthMiddleware(userRepo domain.UserRepository) gin.HandlerFunc {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// ✅ FAIL FAST như trong service
		panic("FATAL: JWT_SECRET is not set in AuthMiddleware")
	}
	// ...
}
```

---

## 2. KIỂM SOÁT DỮ LIỆU ĐẦU VÀO (Input Validation & Sanitization)

### ✅ ĐÃ ĐẢM BẢO

#### 2.1 SQL Injection - Parameterized Queries
**File:** [apps/api/internal/repository/video_repository.go](apps/api/internal/repository/video_repository.go#L256)
```go
if err := r.db.Raw(sql, query, query, limit).Scan(&results).Error; err != nil {
    return nil, err
}
```
- ✅ Sử dụng `?` placeholders thay vì string concatenation
- ✅ GORM tự động escape parameters

**File:** [apps/api/internal/repository/tag_repository_v1.go](apps/api/internal/repository/tag_repository_v1.go#L136)
```go
return r.db.Exec("INSERT INTO video_tags (video_id, tag_id) VALUES (?, ?)", videoID, tagID).Error
```
- ✅ Tất cả queries đều dùng parameterized form

**✅ KHÔNG phát hiện SQL injection vulnerability trong toàn bộ codebase.**

#### 2.2 XSS (Cross-Site Scripting) - React Safe by Default
- ✅ **KHÔNG có `dangerouslySetInnerHTML`** trong toàn bộ codebase React
- ✅ React tự động escape user input khi render
- ✅ Frontend không có innerHTML manipulation trực tiếp

#### 2.3 Input Validation - Gin Binding Tags
**File:** [apps/api/internal/dto/auth.go](apps/api/internal/dto/auth.go#L15-L18)
```go
type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"omitempty,max=100"`
}
```
- ✅ Validate format (email, min length, max length)
- ✅ Required fields được enforce
- ✅ Gin tự động reject invalid requests với 400 Bad Request

---

### 🔴 LỖ HỔNG

#### 🟡 [TRUNG BÌNH] 2.1: Không validate Email uniqueness trước khi signup

**📍 Vị trí:** [apps/api/internal/service/auth_service.go](apps/api/internal/service/auth_service.go#L145-L147)

**🐛 Lỗi bảo mật:**
```go
if _, err := s.userRepo.GetUserByEmail(req.Email); err == nil {
    return nil, errors.New("email already exists")
}
```
Mặc dù có check, nhưng nếu 2 request signup cùng lúc với cùng email, có thể bypass validation (race condition).

**💡 Giải thích:**  
Request 1 và Request 2 cùng check `GetUserByEmail()` → cả 2 đều pass → cả 2 tạo user với cùng email.

**🛠️ Cách sửa:**  
Thêm UNIQUE constraint ở database level:
```sql
-- Migration file
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);
```

---

## 3. BẢO MẬT GIAO THỨC VÀ HEADER (Security Headers & CORS)

### ✅ ĐÃ ĐẢM BẢO

#### 3.1 CORS - Restricted Origin
**File:** [apps/api/internal/middleware/cors.go](apps/api/internal/middleware/cors.go#L11-L24)
```go
func CORS() gin.HandlerFunc {
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	if origin == allowedOrigin || origin == "http://localhost:3000" {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}
```
- ✅ **KHÔNG dùng `*` wildcard**
- ✅ Chỉ cho phép origin được cấu hình
- ✅ Credentials được enable đúng cách

---

### 🔴 LỖ HỔNG NGHIÊM TRỌNG

#### 🔴 [NGHIÊM TRỌNG] 3.1: THIẾU Security Headers

**📍 Vị trí:** Toàn bộ API responses

**🐛 Lỗi bảo mật:**  
API KHÔNG set các security headers quan trọng:
1. **Content-Security-Policy (CSP)** - Chống XSS, code injection
2. **Strict-Transport-Security (HSTS)** - Ép HTTPS
3. **X-Frame-Options** - Chống clickjacking
4. **X-Content-Type-Options** - Chống MIME sniffing

**💡 Giải thích:**  
- **CSP** ngăn trình duyệt load script từ domain lạ
- **HSTS** đảm bảo mọi request đều qua HTTPS
- **X-Frame-Options** ngăn website bị nhúng vào `<iframe>` malicious
- **X-Content-Type-Options: nosniff** ngăn browser đoán MIME type

**🛠️ Cách sửa:**

**Tạo middleware Security Headers:**  
**File mới:** `apps/api/internal/middleware/security_headers.go`
```go
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security-related HTTP headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'; frame-ancestors 'none';")
		
		// Strict Transport Security (HSTS) - 1 year
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent MIME sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// XSS Protection (legacy, but still good practice)
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy (previously Feature-Policy)
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
```

**Apply vào router:**  
**File:** `apps/api/internal/routes/routes.go`
```go
func RegisterRoutes(router *gin.Engine, /* ... */) {
	// Apply security headers to all routes
	router.Use(middleware.SecurityHeaders()) // ✅ Thêm dòng này
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())
	// ...
}
```

**Verify:**
```bash
curl -I http://localhost:8080/api/health

# Expected:
# Strict-Transport-Security: max-age=31536000; includeSubDomains
# X-Frame-Options: DENY
# X-Content-Type-Options: nosniff
# Content-Security-Policy: default-src 'self'; ...
```

---

#### 🟡 [TRUNG BÌNH] 3.2: HTTPS không được enforce trong development

**📍 Vị trí:** [apps/api/internal/handler/auth_handler.go](apps/api/internal/handler/auth_handler.go#L25)

**🐛 Lỗi bảo mật:**
```go
secure := os.Getenv("ENV") == "production"
```
Trong development, cookies được gửi qua HTTP (không mã hóa).

**💡 Giải thích:**  
Developer có thể vô tình test trên network không an toàn → JWT token bị sniff.

**🛠️ Cách sửa:**  
Development nên dùng HTTPS với self-signed certificate hoặc mkcert:
```bash
# Install mkcert
brew install mkcert  # macOS
choco install mkcert # Windows

# Create local CA
mkcert -install

# Generate certificate
cd apps/api
mkcert localhost 127.0.0.1 ::1

# Update Gin to use TLS
# File: apps/api/cmd/api/main.go
server := &http.Server{
    Addr:    ":8443",
    Handler: router,
}
server.ListenAndServeTLS("localhost+2.pem", "localhost+2-key.pem")
```

---

## 4. QUẢN LÝ BÍ MẬT (Secrets Management)

### ✅ ĐÃ ĐẢM BẢO

#### 4.1 Environment Variables - Không commit
- ✅ File `.env` không có trong git history (đã verify)
- ✅ Chỉ có `.env.example` được commit
- ✅ Secrets được load từ environment:
  - `JWT_SECRET`
  - `DB_PASSWORD`
  - `GOOGLE_CLIENT_SECRET`
  - `OPENAI_API_KEY`

#### 4.2 Password Hashing - bcrypt
**File:** [apps/api/internal/service/auth_service.go](apps/api/internal/service/auth_service.go#L150)
```go
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
```
- ✅ Sử dụng bcrypt với `DefaultCost` (cost = 10, ~100ms)
- ✅ Password KHÔNG bao giờ được lưu plaintext

---

### 🔴 LỖ HỔNG

#### 🟡 [TRUNG BÌNH] 4.1: .env.example chứa password mẫu

**📍 Vị trí:** [.env.example](/.env.example#L13)

**🐛 Lỗi bảo mật:**
```dotenv
DB_PASSWORD=ttt_password
```

**💡 Giải thích:**  
Developer có thể copy-paste và quên đổi password, dẫn đến production dùng password mặc định.

**🛠️ Cách sửa:**
```dotenv
# .env.example
DB_PASSWORD=CHANGE_THIS_TO_STRONG_PASSWORD
# Or
DB_PASSWORD=YOUR_SECURE_PASSWORD_HERE
```

---

## 5. LOGGING VÀ MONITORING (Logging & Error Handling)

### ✅ ĐÃ ĐẢM BẢO

#### 5.1 Structured Logging - zerolog
**File:** [apps/api/internal/middleware/logger.go](apps/api/internal/middleware/logger.go#L24-L36)
```go
logEvent := log.Info().
    Str("method", c.Request.Method).
    Str("path", path).
    Int("status", statusCode).
    Dur("latency", latency).
    Str("ip", c.ClientIP())

if len(c.Errors) > 0 {
    logEvent.Str("errors", c.Errors.String())
}
```
- ✅ Structured JSON logging với zerolog
- ✅ Log request method, path, status, latency, IP
- ✅ Có thể dễ dàng parse và phân tích logs

---

### 🔴 LỖ HỔNG NGHIÊM TRỌNG

#### 🔴 [CAO] 5.1: Error Messages Leak Thông Tin Nhạy Cảm

**📍 Vị trí:** Multiple files trong handlers

**🐛 Lỗi bảo mật:**
**File:** [apps/api/internal/handler/auth_handler.go](apps/api/internal/handler/auth_handler.go#L95)
```go
c.JSON(http.StatusBadRequest, dto.ErrorResponse{
    Error:   "Invalid request body",
    Message: err.Error(), // ⚠️ Leak raw validation error
    Code:    http.StatusBadRequest,
})
```

**File:** [apps/api/internal/handler/auth_handler.go](apps/api/internal/handler/auth_handler.go#L112)
```go
c.JSON(statusCode, dto.ErrorResponse{
    Error:   "Authentication failed",
    Message: err.Error(), // ⚠️ Có thể leak "invalid password" vs "user not found"
})
```

**💡 Giải thích:**  
Attacker có thể dựa vào error message để:
1. **Account Enumeration**: Phân biệt "username không tồn tại" vs "password sai"
2. **Database Schema Discovery**: Error như `column 'password_hash' not found`
3. **Version Fingerprinting**: Error từ thư viện → biết version đang dùng

**Ví dụ thực tế:**
```bash
# Request 1
curl -X POST /api/auth/login -d '{"username":"admin","password":"wrong"}'
# Response: "invalid username or password"

# Request 2
curl -X POST /api/auth/login -d '{"username":"nonexist","password":"test"}'
# Response: "invalid username or password"

# ✅ GOOD: Không phân biệt được username có tồn tại hay không
```

**🛠️ Cách sửa:**

**Bước 1:** Tạo helper function cho generic errors:  
**File:** `apps/api/internal/dto/common.go`
```go
package dto

import "github.com/gin-gonic/gin"

// NewInternalErrorResponse returns a generic 500 error without leaking details
func NewInternalErrorResponse(internalMsg string) ErrorResponse {
	// Log the real error internally (for debugging)
	// But return generic message to client
	return ErrorResponse{
		Error:   "Internal Server Error",
		Message: "An unexpected error occurred. Please try again later.",
		Code:    500,
	}
}

// NewBadRequestResponse returns a generic 400 error
func NewBadRequestResponse() ErrorResponse {
	return ErrorResponse{
		Error:   "Bad Request",
		Message: "Invalid request format or parameters",
		Code:    400,
	}
}

// NewUnauthorizedResponse returns a generic 401 error
func NewUnauthorizedResponse() ErrorResponse {
	return ErrorResponse{
		Error:   "Unauthorized",
		Message: "Invalid credentials",
		Code:    401,
	}
}
```

**Bước 2:** Update handlers để dùng generic errors:  
**File:** `apps/api/internal/handler/auth_handler.go`
```go
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// ✅ Log real error internally
		log.Warn().Err(err).Msg("Login request validation failed")
		
		// ✅ Return generic error to client
		c.JSON(http.StatusBadRequest, dto.NewBadRequestResponse())
		return
	}

	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	response, err := h.service.Login(req, userAgent, clientIP)
	if err != nil {
		// ✅ Log real error with context
		log.Warn().
			Err(err).
			Str("username", req.Username).
			Str("ip", clientIP).
			Msg("Login failed")
		
		// ✅ KHÔNG phân biệt "user not found" vs "wrong password"
		c.JSON(http.StatusUnauthorized, dto.NewUnauthorizedResponse())
		return
	}

	// ...
}
```

**Bước 3:** Update service layer để return generic errors:  
**File:** `apps/api/internal/service/auth_service.go`
```go
func (s *authService) Login(req dto.LoginRequest, userAgent, clientIP string) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		// ❌ TRƯỚC: return nil, errors.New("invalid username or password")
		// ✅ SAU: Log internally, return generic error
		return nil, domain.ErrInvalidCredentials // Define constant error
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials // Same error as above
	}
	// ...
}
```

**Bước 4:** Define error constants:  
**File mới:** `apps/api/internal/domain/errors.go`
```go
package domain

import "errors"

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDeactivated = errors.New("account deactivated")
	ErrSessionExpired     = errors.New("session expired")
	
	// Authorization errors
	ErrForbidden = errors.New("forbidden")
	
	// Generic errors (never expose details)
	ErrInternal = errors.New("internal error")
)
```

---

#### 🟡 [TRUNG BÌNH] 5.2: Không log failed login attempts

**📍 Vị trí:** [apps/api/internal/handler/auth_handler.go](apps/api/internal/handler/auth_handler.go#L105-L115)

**🐛 Lỗi bảo mật:**  
Khi login fail, không có log structured để track:
- IP address của attacker
- Số lần thử trong 1 phút
- Pattern của brute-force attack

**💡 Giải thích:**  
Không thể phát hiện và block brute-force attack nếu không có logging.

**🛠️ Cách sửa:**
```go
func (h *AuthHandler) Login(c *gin.Context) {
	// ...
	response, err := h.service.Login(req, userAgent, clientIP)
	if err != nil {
		// ✅ Log failed login với context đầy đủ
		log.Warn().
			Err(err).
			Str("username", req.Username).
			Str("ip", clientIP).
			Str("user_agent", userAgent).
			Msg("Failed login attempt")
		
		c.JSON(http.StatusUnauthorized, dto.NewUnauthorizedResponse())
		return
	}
	// ...
}
```

**Monitoring tip:**  
Setup alert khi có >10 failed logins từ cùng IP trong 5 phút:
```bash
# Example với Prometheus/Grafana
rate(failed_login_total[5m]) > 10
```

---

## 6. ADDITIONAL FINDINGS

### 🟡 [TRUNG BÌNH] 6.1: Middleware auth.go có duplicate logic với service

**📍 Vị trí:**  
- [apps/api/internal/middleware/auth.go](apps/api/internal/middleware/auth.go#L16-L19)
- [apps/api/internal/service/auth_service.go](apps/api/internal/service/auth_service.go#L43-L47)

**🐛 Vấn đề:**  
Cả middleware lẫn service đều load `JWT_SECRET`. Nếu update logic (vd: thêm key rotation), phải sửa 2 nơi.

**🛠️ Khuyến nghị:**  
Centralize JWT logic vào service, middleware chỉ gọi service:
```go
// Middleware
func AuthMiddleware(authService domain.AuthService, userRepo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := getTokenFromRequest(c)
		claims, err := authService.VerifyToken(tokenString) // ✅ Delegate to service
		// ...
	}
}
```

---

### ✅ 6.2: OpenAI API Key được handle an toàn

**File:** [apps/api/internal/server/server.go](apps/api/internal/server/server.go#L56-L59)
```go
openAIClient, err := infrastructure.NewOpenAIClient()
if err != nil {
    log.Warn().Err(err).Msg("Failed to initialize OpenAI client - vector search will be disabled")
    openAIClient = nil // ✅ Graceful degradation
}
```
- ✅ App vẫn chạy được nếu không có OpenAI key
- ✅ Vector search bị disable nhưng core features vẫn hoạt động

---

## 7. CHECKLIST SUMMARY

| Tiêu chí | Trạng thái | Ghi chú |
|----------|-----------|---------|
| **1. Authentication & Authorization** |  |  |
| JWT Secret từ env | ✅ PASS | Service layer có fail-fast |
| HttpOnly Cookies | ✅ PASS | Token không thể bị XSS |
| Secure Cookie (HTTPS) | ✅ PASS | Chỉ trong production |
| Refresh Token Rotation | ✅ PASS | Session-based |
| Session Blacklist | ✅ PASS | Có logout/logoutAll |
| Role-based Access Control | ✅ PASS | RequireRole middleware |
| Rate Limiting | 🔴 FAIL | **NGHIÊM TRỌNG** - Không có |
| Default Secret Fallback | 🟡 WARNING | Middleware có fallback |
| **2. Input Validation** |  |  |
| SQL Injection | ✅ PASS | Tất cả dùng parameterized |
| XSS Prevention | ✅ PASS | React safe, không có dangerouslySetInnerHTML |
| Input Validation | ✅ PASS | Gin binding tags |
| Email Uniqueness | 🟡 WARNING | Race condition possible |
| **3. Security Headers** |  |  |
| CORS Config | ✅ PASS | Restricted origin |
| Content-Security-Policy | 🔴 FAIL | **NGHIÊM TRỌNG** - Thiếu |
| HSTS | 🔴 FAIL | **NGHIÊM TRỌNG** - Thiếu |
| X-Frame-Options | 🔴 FAIL | **NGHIÊM TRỌNG** - Thiếu |
| X-Content-Type-Options | 🔴 FAIL | **NGHIÊM TRỌNG** - Thiếu |
| **4. Secrets Management** |  |  |
| .env not committed | ✅ PASS | Verified git history |
| Environment Variables | ✅ PASS | Tất cả secrets từ env |
| Password Hashing | ✅ PASS | bcrypt DefaultCost |
| .env.example password | 🟡 WARNING | Nên dùng placeholder |
| **5. Logging & Error Handling** |  |  |
| Structured Logging | ✅ PASS | zerolog JSON format |
| Error Message Leakage | 🔴 FAIL | **CAO** - Leak validation errors |
| Failed Login Logging | 🟡 WARNING | Không log đầy đủ context |

---

## 8. PRIORITY ACTION ITEMS

### 🔥 CRITICAL (Phải fix trước khi production)

1. **Implement Rate Limiting**
   - [ ] Install `github.com/ulule/limiter/v3`
   - [ ] Create `RateLimitAuth()` middleware
   - [ ] Apply to `/auth/login` và `/auth/signup`
   - [ ] Test: 6 requests/min → 429 Too Many Requests

2. **Add Security Headers**
   - [ ] Create `SecurityHeaders()` middleware
   - [ ] Apply CSP, HSTS, X-Frame-Options, X-Content-Type-Options
   - [ ] Verify với `curl -I`

3. **Fix Error Message Leakage**
   - [ ] Create generic error helpers
   - [ ] Update all handlers để dùng generic errors
   - [ ] Log real errors internally với `log.Warn()`
   - [ ] Test: Error messages không leak technical details

### ⚠️ HIGH (Nên fix trong sprint tiếp theo)

4. **Remove Default JWT Secret Fallback**
   - [ ] Update `middleware/auth.go` để panic nếu không có secret
   - [ ] Verify: App crash khi `JWT_SECRET` missing

5. **Add Failed Login Logging**
   - [ ] Log IP, username, user-agent cho failed attempts
   - [ ] Setup monitoring/alerting cho brute-force patterns

### 📝 MEDIUM (Technical debt, không blocking)

6. **Email Uniqueness Race Condition**
   - [ ] Add UNIQUE constraint ở database
   - [ ] Migration: `ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email)`

7. **Centralize JWT Logic**
   - [ ] Refactor middleware để dùng `authService.VerifyToken()`
   - [ ] Remove duplicate JWT secret loading

8. **Update .env.example**
   - [ ] Change `DB_PASSWORD=ttt_password` → `DB_PASSWORD=YOUR_SECURE_PASSWORD_HERE`

---

## 9. CONCLUSION

**Tổng quan:**  
Project TTT có foundation bảo mật tốt với JWT, bcrypt, parameterized queries, và CORS config. Tuy nhiên, còn thiếu các lớp bảo vệ quan trọng cho production:

**Điểm cần cải thiện:**
- **Rate Limiting** để chống brute-force
- **Security Headers** để tăng defense-in-depth
- **Error Handling** để không leak thông tin

**Khuyến nghị:**  
Fix 3 critical issues trên trước khi deploy production. Nếu không, risk bị tấn công:
- Brute-force login (không có rate limit)
- XSS/Clickjacking (không có CSP/X-Frame-Options)
- Information disclosure (error messages leak)

**Timeline đề xuất:**
- **Sprint hiện tại (Week 1-2):** Fix Critical issues (rate limit, security headers, error handling)
- **Sprint tiếp theo (Week 3-4):** Fix High priority issues
- **Backlog:** Medium priority issues (technical debt)

---

**Người thực hiện:** GitHub Copilot Security Audit  
**Ngày:** 18/12/2025  
**Version:** 1.0
