# Agents - Multi-Provider AI Automation Framework

## 🎯 Tổng quan

Framework automation với hỗ trợ **nhiều AI providers** sử dụng **official SDKs**:
- ✅ Google Gemini (via `google-generativeai`)
- ✅ OpenAI (via `openai`)
- 🔧 Dễ dàng mở rộng thêm provider khác

## 📁 Kiến trúc mới

```
agents/
├── base/                    # Base classes & interfaces
│   ├── __init__.py
│   └── provider.py         # BaseLLMProvider interface
│
├── llm/                     # LLM providers implementations
│   ├── __init__.py
│   ├── gemini.py           # Google Gemini provider (official SDK)
│   ├── openai.py           # OpenAI provider (official SDK)
│   └── manager.py          # LLM Manager (factory)
│
├── utils/                   # Utilities
│   ├── __init__.py
│   ├── optimizer.py        # Context optimizer, JSON parser
│   ├── logger.py           # Structured logging
│   └── update_json_start_time.py  # Fix tool for missing start_time fields
│
├── workflows/               # Automation workflows
│   ├── __init__.py
│   └── transcript_to_json.py  # Transcript to JSON workflow
│
├── prompts/                 # Prompt templates
│   └── transcript_to_json_instruction.md
│
├── main.py                  # Entry point
├── test_providers.py        # Test suite
├── .env                     # Configuration (gitignored)
├── .env.example             # Configuration template
├── requirements.txt         # Dependencies
└── README.md               # This file
```

## 🔄 Flow Diagram

```
┌──────────────┐
│   main.py    │  ← Entry point
└──────┬───────┘
       │
┌──────▼────────────────────┐
│  workflows/               │
│  transcript_to_json.py    │  ← Workflow logic
└──────┬────────────────────┘
       │
┌──────▼────────────────────┐
│  llm/manager.py           │  ← Factory & provider manager
│  (LLMManager)             │
└──────┬────────────────────┘
       │
┌──────▼─────────────────────────────────┐
│  llm/                                   │
│  ┌──────────┐     ┌──────────┐        │
│  │gemini.py │     │openai.py │ ...    │  ← Provider implementations
│  └──────────┘     └──────────┘        │     (using official SDKs)
└─────────────────────────────────────────┘
       │
┌──────▼────────────────────┐
│  base/provider.py         │  ← Base interface
│  (BaseLLMProvider)        │
└───────────────────────────┘
```

## 🚀 Thiết lập

### 1. Cài đặt dependencies

```bash
pip install -r requirements.txt
```

**Dependencies:**
- `google-genai` - Google Gemini official SDK (latest)
- `openai` - OpenAI official SDK  
- `python-dotenv` - Load .env files

### 2. Cấu hình API Keys

Sao chép file mẫu:
```bash
cp .env.example .env
```

Chỉnh sửa `.env`:
```bash
# Chọn provider mặc định
DEFAULT_AI_PROVIDER=gemini

# Gemini configuration
GEMINI_API_KEYS=key1,key2,key3
GEMINI_MODEL=gemini-2.0-flash-exp

# OpenAI configuration
OPENAI_API_KEYS=sk-proj-key1,sk-proj-key2
OPENAI_MODEL=gpt-4o-mini

# Custom Base URL cho dịch vụ OpenAI bên thứ 3 (optional)
# Ví dụ: https://v98store.com/v1
# Nếu không set, dùng API gốc của OpenAI
OPENAI_BASE_URL=
```

**Lưu ý:** Nếu bạn sử dụng dịch vụ OpenAI bên thứ 3, hãy set `OPENAI_BASE_URL`. Ví dụ:

```bash
OPENAI_BASE_URL=https://v98store.com/v1
```

API call format sẽ là:
```
POST https://v98store.com/v1/chat/completions
Authorization: Bearer YOUR_API_KEY
```

### 3. Chạy workflow

```bash
# Dùng default provider
python main.py

# Chỉ định provider cụ thể
python main.py gemini
python main.py openai
```

## 🛠 Utilities

### Cập nhật start_time cho JSON output

Trong trường hợp các tệp JSON output thiếu trường `start_time` (cần thiết cho quá trình import vào database), bạn có thể sử dụng công cụ sau để bổ sung giá trị mặc định (`0`):

```bash
python agents/utils/update_json_start_time.py
```

Công cụ này sẽ quét thư mục `agents/resources/transcript_to_json/output` và cập nhật tất cả các tệp JSON có cấu trúc `transcript`.

## 4. Chạy tests

```bash
python test_providers.py
```

## 💡 Sử dụng

### Chạy workflow có sẵn

```python
import asyncio
from workflows import run_transcript_to_json

# Run với default provider
asyncio.run(run_transcript_to_json())

# Run với provider cụ thể
asyncio.run(run_transcript_to_json(provider="openai"))
```

### Sử dụng LLM Manager trực tiếp

```python
from llm.manager import llm_manager

# Get default provider
provider = llm_manager.get_provider()

# Generate text
result = await provider.generate(
    prompt="Your prompt",
    system_instruction="System instruction"
)

# Get specific provider
gemini = llm_manager.get_provider("gemini")
openai = llm_manager.get_provider("openai")
```

### Tạo workflow mới

```python
# workflows/your_workflow.py
import asyncio
from llm.manager import llm_manager
from utils import ContextOptimizer

async def your_workflow():
    # Get LLM provider
    llm = llm_manager.get_provider()
    
    # Your logic here
    result = await llm.generate(
        prompt="Your prompt",
        system_instruction="Your instruction"
    )
    
    # Process result
    data = ContextOptimizer.parse_json_safely(result)
    
    return data

if __name__ == "__main__":
    asyncio.run(your_workflow())
```

## Cấu hình có thể tùy chỉnh

### config_manager.py
- `DEFAULT_PROVIDER` - Provider mặc định ("gemini", "openai")
- `DEFAULT_GEMINI_MODEL` - Model Gemini mặc định
- `DEFAULT_OPENAI_MODEL` - Model OpenAI mặc định
- `MAX_RETRIES` - Số lần retry tối đa
- `RETRY_BASE_MS` - Thời gian delay cơ bản (ms)

### llm_providers.py
- `CONNECTOR_LIMIT` - Số kết nối TCP tối đa
- `KEEPALIVE_TIMEOUT` - Timeout keepalive (giây)
- `TEMPERATURE` - Temperature cho AI (0.0-1.0)
- `MAX_OUTPUT_TOKENS` - Số token tối đa output
- `JITTER_MAX_MS` - Random jitter tối đa (ms)

### main.py
- `INPUT_DIR` - Thư mục chứa file input
- `MAX_CONCURRENT_WORKERS` - Số worker chạy đồng thời
- `INPUT_FILE_EXT` - Extension file input
- `OUTPUT_FILE_EXT` - Extension file output
- `ENCODING` - Encoding file

### optimization_utils.py
- `TOKEN_ESTIMATE_DIVISOR` - Ước lượng token
- `MAX_CONTENT_CHARS` - Độ dài content tối đa

## Workflow

1. **Input**: Đọc file `.f.txt` từ `INPUT_DIR`
2. **Optimize**: Cắt bớt nội dung nếu quá dài
3. **Process**: Gọi AI API (Gemini/OpenAI) để phân tích transcript
4. **Validate**: Kiểm tra JSON output hợp lệ
5. **Output**: Lưu kết quả vào file `.json`

## Features

### ✅ Multi-Provider Support
- Hỗ trợ nhiều AI providers (Gemini, OpenAI)
- Dễ dàng mở rộng thêm provider mới
- Chuyển đổi provider runtime

### ✅ Intelligent Retry Logic
- Exponential backoff với jitter
- Round-robin API key rotation
- Tự động chuyển key khi rate limit

### ✅ Error Handling
- Network error → Retry với delay
- Rate limit → Chuyển key + exponential backoff
- Invalid JSON → Lưu raw output để debug

## Troubleshooting

### Lỗi "Không tìm thấy API key"
```bash
# Kiểm tra file .env có tồn tại không
ls .env

# Kiểm tra nội dung
cat .env

# Đảm bảo có ít nhất một provider được cấu hình
```

### Lỗi "Provider không được implement"
```bash
# Kiểm tra tên provider trong .env
# Phải là: "gemini" hoặc "openai" (lowercase)
DEFAULT_AI_PROVIDER=gemini
```

### Rate limit quá nhanh
```bash
# Giảm MAX_CONCURRENT_WORKERS trong main.py
MAX_CONCURRENT_WORKERS = 2

# Hoặc thêm nhiều API key hơn
GEMINI_API_KEYS=key1,key2,key3,key4,key5
```
