/**
 * Video Type Definitions
 * Matching backend DTOs from apps/api/internal/dto
 */

import type { TagResponse as ModTagResponse } from './tag'

// ===== API Request Types =====

export type VideoSort = 'newest' | 'popular' | 'views'

export interface ListVideoRequest {
  page?: number // Default: 1, min: 1
  limit?: number // Default: 10, min: 1, max: 50
  sort?: VideoSort
  tag_id?: string // UUID
  has_transcript?: boolean // Filter by transcript: true = only with, false = only without, undefined = all
}

// VideoListRequest for mod dashboard
export interface VideoListRequest {
  page?: number
  page_size?: number
  q?: string
  tag_ids?: number[]
}

export interface TranscriptSearchRequest {
  q: string // min length: 2
  limit?: number // Default: 20, min: 1, max: 50
}

export interface TagSearchRequest {
  q: string // min length: 2
  limit?: number // Default: 5, min: 1, max: 10
}

// ===== API Response Types =====

export interface VideoCardResponse {
  id: string
  youtube_id: string
  title: string
  thumbnail_url: string
  duration: number // Seconds
  published_at: string
  view_count: number
  has_transcript: boolean // Show "CC" badge
  review_count: number // Number of reviews - show "Đã duyệt" badge if > 0
}

export interface TagResponse {
  id: string
  name: string
}

export interface VideoDetailResponse extends VideoCardResponse {
  tags: TagResponse[]
}

export interface SegmentResponse {
  id: number
  start_time: number // Milliseconds
  end_time: number // Milliseconds
  text: string
}

export interface TranscriptResponse {
  video_id: string
  segments: SegmentResponse[]
}

export interface PaginationMetadata {
  page: number
  limit: number
  total_items: number
  total_pages: number
}

export interface VideoListResponse {
  data: VideoCardResponse[]
  pagination: PaginationMetadata
}

export interface TranscriptSearchResult {
  video_id: string
  video_title: string
  thumbnail_url: string
  start_time: number
  end_time: number
  text: string
  rank: number // Relevance score
}

export interface TranscriptSearchResponse {
  query: string
  results: TranscriptSearchResult[]
  total: number
}

export interface TagSearchResult {
  id: string
  name: string
  similarity: number // Cosine similarity (0-1)
}

export interface TagSearchResponse {
  query: string
  results: TagSearchResult[]
  total: number
}

export interface ErrorResponse {
  error: string
  message?: string
  code: number
}

// ===== Video Transcript Review Types =====

export interface SubmitReviewRequest {
  notes?: string // Optional review notes (max 500 chars)
}

export interface VideoTranscriptReviewResponse {
  id: number
  video_id: string
  user_id: string
  reviewed_at: string // ISO 8601 timestamp
  total_reviews: number // Total reviews for this video
  video_status: string // Current video status (e.g., "PUBLISHED", "DRAFT")
  points_awarded: number // Points given to reviewer
  message: string // Human-readable status message
}

export interface VideoReviewStats {
  video_id: string
  review_count: number
}

export interface UserReviewStatus {
  video_id: string
  has_reviewed: boolean
}

// ===== Frontend Helper Types =====

export interface TranscriptParagraph {
  id: string
  startTime: number
  segments: SegmentResponse[]
}

// ===== Mod Dashboard Types =====

/**
 * Video model for mod dashboard
 */
export interface Video {
  id: number
  youtube_id: string
  title: string
  description?: string
  thumbnail_url?: string
  duration?: number
  published_at?: string
  view_count?: number
  has_transcript?: boolean
  tags?: ModTagResponse[]
  created_at: string
  updated_at: string
}

/**
 * VideoListResponse for mod dashboard (different structure from public API)
 */
export interface ModVideoListResponse {
  videos: Video[]
  total: number
  page: number
  page_size: number
}
