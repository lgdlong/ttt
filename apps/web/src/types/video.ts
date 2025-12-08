/**
 * Video Type Definitions
 * Matching backend DTOs from apps/api/internal/dto
 */

// ===== API Request Types =====

export type VideoSort = 'newest' | 'popular' | 'views'

export interface ListVideoRequest {
  page?: number // Default: 1, min: 1
  limit?: number // Default: 10, min: 1, max: 50
  sort?: VideoSort
  tag_id?: string // UUID
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

// ===== Frontend Helper Types =====

export interface TranscriptParagraph {
  id: string
  startTime: number
  segments: SegmentResponse[]
}
