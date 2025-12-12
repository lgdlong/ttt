-- Migration: Create HNSW index for vector similarity search on tags
-- Purpose: Optimize semantic search performance using pgvector
-- Requires: pgvector extension enabled

-- Enable pgvector extension if not already enabled
CREATE EXTENSION IF NOT EXISTS vector;

-- Create HNSW index for fast approximate nearest neighbor search
-- Using cosine distance (vector_cosine_ops) as it's best for text embeddings
-- Embeddings: OpenAI text-embedding-3-small (1536 dimensions, native fit, no truncation)
-- HNSW parameters:
--   m = 16: number of bi-directional links (default, good balance)
--   ef_construction = 64: quality of index construction (higher = better recall but slower build)
CREATE INDEX IF NOT EXISTS idx_tags_embedding_hnsw 
ON public.tags 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Performance tuning for vector search queries
-- These settings apply at session level when executing searches
-- ef_search controls the quality of search (higher = better recall but slower)
COMMENT ON INDEX idx_tags_embedding_hnsw IS 'HNSW index for semantic tag search using OpenAI text-embedding-3-small (1536 dimensions, native fit without truncation). Use SET hnsw.ef_search = 100 before queries for better recall.';

-- Verification query to check index exists
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'tags' AND indexname LIKE '%embedding%';
