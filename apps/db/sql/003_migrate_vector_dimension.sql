-- Migration: Update vector dimension for text-embedding-3-small (1536 dims)
-- Purpose: Use smaller embedding model without truncation
-- Reason: text-embedding-3-small fits natively within pgvector; no truncation needed

-- First, drop the dependent HNSW index
DROP INDEX IF EXISTS idx_tags_embedding_hnsw;

-- Drop the embedding column and recreate with 1536 dimension
ALTER TABLE public.tags
DROP COLUMN IF EXISTS embedding;

ALTER TABLE public.tags
ADD COLUMN embedding vector(1536);

-- Recreate HNSW index with new dimension
CREATE INDEX IF NOT EXISTS idx_tags_embedding_hnsw 
ON public.tags 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Update comment
COMMENT ON INDEX idx_tags_embedding_hnsw IS 'HNSW index for semantic tag search using OpenAI text-embedding-3-small (1536 dims). Use SET hnsw.ef_search = 100 before queries for better recall.';

-- Note: After running this migration, re-generate embeddings for all tags
-- using the backfill_embeddings script. Embeddings are 1536 dimensions (no truncation).
