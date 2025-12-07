# Swagger Documentation

## 📚 Đã Cài Đặt

### Dependencies
```bash
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
go install github.com/swaggo/swag/cmd/swag@latest
```

## 🎯 Swagger Annotations

Đã thêm Swagger annotations cho tất cả endpoints:

### Videos Endpoints
- `GET /api/v1/videos` - List videos với pagination, filtering, sorting
- `GET /api/v1/videos/{id}` - Get video detail by UUID
- `GET /api/v1/videos/{id}/transcript` - Get transcript segments

### Search Endpoints
- `GET /api/v1/search/transcript` - Full-text search trong transcripts
- `GET /api/v1/search/tags` - Semantic search với vector embeddings

### System Endpoints
- `GET /api/v1/health` - Health check

## 🚀 Sử Dụng

### 1. Generate Swagger Docs

```bash
# Từ apps/api/
swag init -g cmd/api/main.go -o ./docs

# Hoặc dùng Makefile
make swagger
```

### 2. Start API Server

```bash
pnpm dev:api
# hoặc
go run cmd/api/main.go
```

### 3. Truy Cập Swagger UI

Mở browser và vào:
```
http://localhost:8080/swagger/index.html
```

## 📝 Generated Files

```
apps/api/docs/
├── docs.go         # Go package với embedded docs
├── swagger.json    # OpenAPI 3.0 JSON spec
└── swagger.yaml    # OpenAPI 3.0 YAML spec
```

## 🔧 Configuration

### main.go Header
```go
// @title TTT Video API
// @version 1.0
// @description API for managing YouTube videos, transcripts, and semantic search

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
```

### Example Annotation
```go
// GetVideoList godoc
// @Summary List videos with pagination
// @Description Get a paginated list of videos with optional filtering and sorting
// @Tags Videos
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)" default(1)
// @Param limit query int false "Items per page (default: 20, max: 100)" default(20)
// @Success 200 {object} dto.VideoListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /videos [get]
func (h *VideoHandler) GetVideoList(c *gin.Context) { ... }
```

## 📖 Swagger UI Features

- **Interactive API Testing** - Test endpoints trực tiếp từ UI
- **Request/Response Examples** - Xem example JSON cho từng endpoint
- **Model Schemas** - Explore DTO structures
- **Parameter Documentation** - Chi tiết về query params, path params
- **Response Codes** - Danh sách tất cả possible response codes

## 🔄 Auto-Regenerate

Khi thay đổi annotations:

```bash
# Run lại swag init
make swagger

# Restart server để load docs mới
pnpm dev:api
```

## 💡 Tips

1. **Query Parameters**: Dùng `@Param name query type required "description" default(value)`
2. **Path Parameters**: Dùng `@Param id path string true "ID"`
3. **Request Body**: Dùng `@Param body body dto.Request true "Body"`
4. **Response**: Dùng `@Success 200 {object} dto.Response`
5. **Tags**: Group endpoints bằng `@Tags GroupName`

## 🔗 Resources

- Swagger UI: `http://localhost:8080/swagger/index.html`
- JSON Spec: `http://localhost:8080/swagger/doc.json`
- YAML Spec: Available in `docs/swagger.yaml`

---
**Status**: ✅ Swagger documentation hoàn chỉnh và ready to use
