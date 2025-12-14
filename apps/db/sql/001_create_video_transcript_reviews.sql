-- Migration: Create video_transcript_reviews table
-- Purpose: Track which users have reviewed which videos for moderator KPI and video verification

CREATE TABLE IF NOT EXISTS public.video_transcript_reviews (
    id bigserial PRIMARY KEY,
    video_id uuid NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reviewed_at timestamptz NOT NULL DEFAULT now(),
    
    -- Ensure each user can only review a video once
    CONSTRAINT unique_review UNIQUE (video_id, user_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_video_transcript_reviews_video_id ON public.video_transcript_reviews(video_id);
CREATE INDEX IF NOT EXISTS idx_video_transcript_reviews_user_id ON public.video_transcript_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_video_transcript_reviews_reviewed_at ON public.video_transcript_reviews(reviewed_at DESC);

-- Add comment for documentation
COMMENT ON TABLE public.video_transcript_reviews IS 'Tracks moderator reviews of video transcripts for verification workflow';
COMMENT ON COLUMN public.video_transcript_reviews.video_id IS 'Reference to the video being reviewed';
COMMENT ON COLUMN public.video_transcript_reviews.user_id IS 'Moderator who performed the review';
COMMENT ON COLUMN public.video_transcript_reviews.reviewed_at IS 'Timestamp when review was completed';
