-- Migration: Convert single tags table to Canonical-Alias architecture
-- Purpose: Enable AI-powered semantic deduplication with zero-error flow
-- Strategy: Auto-merge semantically similar tags (e.g., "Money" ← "Tiền", "Cash")

-- Dependencies
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm; -- For text search optimization

-- ============================================================
-- STEP 1: Create Canonical Tags Table (Tag Concepts)
-- ============================================================
CREATE TABLE IF NOT EXISTS canonical_tags (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    
    -- URL-friendly slug (e.g., "money-finance")
    slug VARCHAR(100) UNIQUE NOT NULL,
    
    -- Official display name (e.g., "Money")
    display_name VARCHAR(100) NOT NULL,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for slug lookups (used in URLs)
CREATE INDEX IF NOT EXISTS idx_canonical_tags_slug ON canonical_tags(slug);

-- Index for display_name search
CREATE INDEX IF NOT EXISTS idx_canonical_tags_display_name ON canonical_tags(display_name);

-- ============================================================
-- STEP 2: Create Tag Aliases Table (User Input Variants)
-- ============================================================
CREATE TABLE IF NOT EXISTS tag_aliases (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    
    -- Foreign key to canonical tag (CASCADE delete)
    canonical_tag_id UUID NOT NULL REFERENCES canonical_tags(id) ON DELETE CASCADE,
    
    -- Original text user entered (e.g., "  Tiền  ")
    raw_text VARCHAR(100) NOT NULL,
    
    -- Normalized text for exact matching (e.g., "tiền")
    -- Normalization: LOWER(TRIM(raw_text))
    normalized_text VARCHAR(100) NOT NULL,
    
    -- Language code (e.g., 'vi', 'en', 'zh', 'unk')
    -- Used for language-aware suggestions
    language VARCHAR(10) DEFAULT 'unk',
    
    -- Embedding vector (1536 dimensions from text-embedding-3-small)
    -- No truncation needed - small model fits natively
    embedding vector(1536),
    
    -- Admin review flag (FALSE = AI auto-mapped, needs review)
    is_reviewed BOOLEAN DEFAULT FALSE,
    
    -- Similarity score when mapped (1.0 = exact canonical, 0.85-0.99 = AI mapped)
    similarity_score FLOAT DEFAULT 1.0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================
-- STEP 3: Create Indexes for Performance
-- ============================================================

-- Index 1: Fast exact match (Layer 1 - Cache Hit)
-- Used for: SELECT * FROM tag_aliases WHERE normalized_text = 'tiền'
CREATE INDEX IF NOT EXISTS idx_tag_aliases_normalized 
ON tag_aliases(normalized_text);

-- Index 2: Canonical tag lookup (used in JOINs)
CREATE INDEX IF NOT EXISTS idx_tag_aliases_canonical_tag_id 
ON tag_aliases(canonical_tag_id);

-- Index 3: Vector similarity search (Layer 3 - Semantic Search)
-- HNSW = Hierarchical Navigable Small World (fast approximate nearest neighbor)
-- Parameters:
--   m = 16: connectivity (default, good balance)
--   ef_construction = 64: build quality (higher = slower build but better recall)
CREATE INDEX IF NOT EXISTS idx_tag_aliases_embedding_hnsw 
ON tag_aliases 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Index 4: Trigram index for fuzzy text search (optional, for autocomplete)
CREATE INDEX IF NOT EXISTS idx_tag_aliases_raw_text_trgm 
ON tag_aliases 
USING gin (raw_text gin_trgm_ops);

-- ============================================================
-- STEP 4: Update Video-Tag Relationship Table
-- ============================================================

-- Rename old video_tags table (backup)
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'video_tags') THEN
        ALTER TABLE video_tags RENAME TO video_tags_old_backup;
    END IF;
END $$;

-- Create new video_tags table pointing to canonical tags
CREATE TABLE IF NOT EXISTS video_tags (
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    canonical_tag_id UUID NOT NULL REFERENCES canonical_tags(id) ON DELETE CASCADE,
    
    -- Timestamp for tracking when tag was added
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Composite primary key (one tag per video)
    PRIMARY KEY (video_id, canonical_tag_id)
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_video_tags_video_id ON video_tags(video_id);
CREATE INDEX IF NOT EXISTS idx_video_tags_canonical_tag_id ON video_tags(canonical_tag_id);

-- ============================================================
-- STEP 5: Add Constraints and Triggers
-- ============================================================

-- Constraint: normalized_text must be unique per canonical tag
-- This prevents duplicate aliases like "money" appearing twice under same canonical
CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_aliases_normalized_unique 
ON tag_aliases(normalized_text);

-- Trigger: Auto-update updated_at on canonical_tags
CREATE OR REPLACE FUNCTION update_canonical_tags_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_canonical_tags_updated_at ON canonical_tags;
CREATE TRIGGER trigger_canonical_tags_updated_at
    BEFORE UPDATE ON canonical_tags
    FOR EACH ROW
    EXECUTE FUNCTION update_canonical_tags_updated_at();

-- ============================================================
-- STEP 6: Add Comments for Documentation
-- ============================================================

COMMENT ON TABLE canonical_tags IS 'Master tag concepts (e.g., "Money"). Each represents a unique semantic concept.';
COMMENT ON TABLE tag_aliases IS 'User input variants (e.g., "Tiền", "Cash", "Money") that map to canonical tags. Contains embeddings for semantic search.';
COMMENT ON TABLE video_tags IS 'Many-to-many relationship between videos and canonical tags.';

COMMENT ON COLUMN tag_aliases.normalized_text IS 'LOWER(TRIM(raw_text)) for fast exact matching. Used in Layer 1 cache hit optimization.';
COMMENT ON COLUMN tag_aliases.embedding IS 'OpenAI text-embedding-3-small (1536 dims). Used in Layer 3 semantic search.';
COMMENT ON COLUMN tag_aliases.similarity_score IS 'Confidence score: 1.0 = exact canonical, 0.85-0.99 = AI auto-mapped, <0.85 = manual override.';
COMMENT ON COLUMN tag_aliases.is_reviewed IS 'FALSE = AI auto-mapped (needs admin review), TRUE = human verified or canonical.';

COMMENT ON INDEX idx_tag_aliases_embedding_hnsw IS 'Vector similarity search index. Use SET hnsw.ef_search = 100 before queries for better recall.';

-- ============================================================
-- VERIFICATION QUERIES
-- ============================================================

-- Check tables exist
-- SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%tag%';

-- Check indexes
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename IN ('canonical_tags', 'tag_aliases', 'video_tags');

-- Check constraints
-- SELECT conname, contype FROM pg_constraint WHERE conrelid IN ('canonical_tags'::regclass, 'tag_aliases'::regclass);
