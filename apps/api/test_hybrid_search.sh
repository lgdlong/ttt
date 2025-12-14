#!/usr/bin/env bash
# Test Hybrid Search Feature
# Usage: ./test_hybrid_search.sh

API_BASE="http://localhost:8080"

echo "🧪 Testing Hybrid Search Implementation"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test 1: SQL Search (Free)
echo -e "${YELLOW}Test 1: SQL LIKE Search (Miễn phí)${NC}"
echo "Query: 'Java'"
echo ""
curl -s "$API_BASE/mod/tags?query=Java&limit=5" | jq .
echo ""
echo -e "${GREEN}✅ Expected: Tìm thấy bằng SQL LIKE, không gọi OpenAI${NC}"
echo -e "${GREEN}💰 Cost: \$0${NC}"
echo ""
echo "---"
echo ""

# Test 2: Create Tag with Embedding
echo -e "${YELLOW}Test 2: Tạo Tag Mới (Auto-generate Embedding)${NC}"
echo "Creating tag: 'Rust'"
echo ""
curl -s -X POST "$API_BASE/mod/tags" \
  -H "Content-Type: application/json" \
  -d '{"name": "Rust"}' | jq .
echo ""
echo -e "${GREEN}✅ Expected: Tag được tạo với embedding tự động${NC}"
echo -e "${GREEN}💰 Cost: ~\$0.00002 (1 OpenAI call)${NC}"
echo ""
echo "---"
echo ""

# Test 3: Vector Search (AI Semantic)
echo -e "${YELLOW}Test 3: Vector Search - Semantic Matching${NC}"
echo "Query: 'ngôn ngữ con rắn' (should find 'Python')"
echo ""
curl -s "$API_BASE/mod/tags?query=ngôn+ngữ+con+rắn&limit=5" | jq .
echo ""
echo -e "${GREEN}✅ Expected: SQL không tìm thấy → Vector search tìm thấy 'Python'${NC}"
echo -e "${GREEN}💰 Cost: ~\$0.00002 (1 OpenAI call)${NC}"
echo ""
echo "---"
echo ""

# Test 4: Another Semantic Search
echo -e "${YELLOW}Test 4: Vector Search - 'programming language for web'${NC}"
echo "Query: 'programming language for web'"
echo ""
curl -s "$API_BASE/mod/tags?query=programming+language+for+web&limit=5" | jq .
echo ""
echo -e "${GREEN}✅ Expected: Find 'JavaScript', 'TypeScript', etc.${NC}"
echo -e "${GREEN}💰 Cost: ~\$0.00002 (1 OpenAI call)${NC}"
echo ""
echo "---"
echo ""

# Test 5: Add Tag to Video (Create if not exists)
echo -e "${YELLOW}Test 5: Add Tag to Video (với auto-create)${NC}"
echo "Note: Replace VIDEO_ID with actual video ID"
echo ""
echo "Command:"
echo 'curl -X POST "$API_BASE/mod/videos/{VIDEO_ID}/tags" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d '"'"'{"tag_name": "Machine Learning"}'"'"
echo ""
echo -e "${GREEN}✅ Expected: Tạo tag 'Machine Learning' với embedding và link với video${NC}"
echo -e "${GREEN}💰 Cost: ~\$0.00002 (1 OpenAI call)${NC}"
echo ""

echo "========================================"
echo -e "${GREEN}✅ All tests completed!${NC}"
echo ""
echo "💡 Tips:"
echo "  - Check server logs để xem SQL/Vector search decision"
echo "  - Monitor OpenAI usage tại: https://platform.openai.com/usage"
echo "  - Majority searches should hit SQL (FREE)"
