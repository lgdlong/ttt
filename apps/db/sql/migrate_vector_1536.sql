-- Emergency Migration: Update vector dimension from 2000 to 1536
-- Reason: Switched from text-embedding-3-large (truncated to 2000) to text-embedding-3-small (native 1536)
-- This fixes: ERROR: expected 2000 dimensions, not 1536

-- Step 1: Drop the dependent HNSW index on tag_aliases
DROP INDEX IF EXISTS idx_tag_aliases_embedding_hnsw;

-- Step 2: Drop the embedding column on tag_aliases and recreate with 1536 dimension
ALTER TABLE public.tag_aliases
DROP COLUMN embedding;

ALTER TABLE public.tag_aliases
ADD COLUMN embedding vector(1536);

-- Step 3: Recreate HNSW index with new dimension on tag_aliases
CREATE INDEX idx_tag_aliases_embedding_hnsw 
ON public.tag_aliases 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Step 4: Update comment
COMMENT ON INDEX idx_tag_aliases_embedding_hnsw IS 'HNSW index for semantic tag search using OpenAI text-embedding-3-small (1536 dims). Use SET hnsw.ef_search = 100 before queries for better recall.';

-- Step 5: Drop the dependent HNSW index on tags (legacy table)
DROP INDEX IF EXISTS idx_tags_embedding_hnsw;

-- Step 6: Update tags table embedding column
ALTER TABLE public.tags
DROP COLUMN IF EXISTS embedding;

ALTER TABLE public.tags
ADD COLUMN embedding vector(1536);

-- Step 7: Recreate HNSW index with new dimension on tags
CREATE INDEX idx_tags_embedding_hnsw 
ON public.tags 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Step 8: Update comment
COMMENT ON INDEX idx_tags_embedding_hnsw IS 'HNSW index for semantic tag search using OpenAI text-embedding-3-small (1536 dims). Use SET hnsw.ef_search = 100 before queries for better recall.';

SELECT 'Migration complete: All embedding columns updated to vector(1536)' as status;
