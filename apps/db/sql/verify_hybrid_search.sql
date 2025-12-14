-- Verify Hybrid Search Setup
-- Run this to check if everything is configured correctly

-- 1. Check if pgvector extension is installed
SELECT * FROM pg_extension WHERE extname = 'vector';
-- Expected: 1 row with vector extension

-- 2. Check if HNSW index exists on tags.embedding
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename = 'tags' 
  AND indexname LIKE '%embedding%';
-- Expected: 1 row with idx_tags_embedding_hnsw

-- 3. Check tag table structure
SELECT 
    column_name,
    data_type,
    character_maximum_length,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'tags'
ORDER BY ordinal_position;
-- Expected: id, name, embedding (vector)

-- 4. Count tags with/without embeddings
SELECT 
    COUNT(*) as total_tags,
    COUNT(embedding) as tags_with_embedding,
    COUNT(*) - COUNT(embedding) as tags_without_embedding
FROM tags;
-- Note: Các tag cũ có thể chưa có embedding

-- 5. Sample tags với embedding info
SELECT 
    id,
    name,
    CASE 
        WHEN embedding IS NOT NULL THEN 'YES'
        ELSE 'NO'
    END as has_embedding,
    CASE 
        WHEN embedding IS NOT NULL THEN array_length(embedding::vector::float4[], 1)
        ELSE NULL
    END as embedding_dimensions
FROM tags 
LIMIT 10;
-- Expected: Các tag mới phải có embedding với 1536 dimensions

-- 6. Test vector search (nếu có data)
-- Replace [0.1, 0.2, ...] với actual vector hoặc skip test này
-- SELECT 
--     name,
--     embedding <=> '[0.1, 0.2, ..., 0.1]'::vector as distance
-- FROM tags
-- WHERE embedding IS NOT NULL
-- ORDER BY distance
-- LIMIT 5;

-- 7. Check index usage statistics (after some queries)
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes
WHERE tablename = 'tags';

-- 8. Performance: Set optimal HNSW search quality
-- Run this before executing vector searches
SET hnsw.ef_search = 100;  -- Higher = better recall but slower (default: 40)

-- 9. Clean up old tags without embeddings (optional)
-- UPDATE tags 
-- SET embedding = NULL 
-- WHERE embedding IS NULL;

COMMENT ON INDEX idx_tags_embedding_hnsw IS 
'HNSW index for hybrid search: SQL LIKE first, then vector search if no results';
