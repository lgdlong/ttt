import { v1ApiClient } from '~/lib/apiClient'
import type {
  ListVideoRequest,
  VideoListResponse,
  VideoDetailResponse,
  TranscriptResponse,
  TranscriptSearchRequest,
  TranscriptSearchResponse,
  SegmentResponse,
  SubmitReviewRequest,
  VideoTranscriptReviewResponse,
} from '~/types/video'

/**
 * Video API Service
 * Handles all video-related API calls to the v1 backend
 */

/**
 * Fetch videos with pagination and filters
 * GET /api/v1/videos
 */
export async function fetchVideos(params?: ListVideoRequest): Promise<VideoListResponse> {
  const response = await v1ApiClient.get<VideoListResponse>('/videos', { params })
  return response.data
}

/**
 * Fetch single video detail by ID
 * GET /api/v1/videos/:id
 */
export async function fetchVideoById(id: string): Promise<VideoDetailResponse> {
  const response = await v1ApiClient.get<VideoDetailResponse>(`/videos/${id}`)
  return response.data
}

/**
 * Fetch video transcript by video ID
 * GET /api/v1/videos/:id/transcript
 */
export async function fetchVideoTranscript(videoId: string): Promise<TranscriptResponse> {
  const response = await v1ApiClient.get<TranscriptResponse>(`/videos/${videoId}/transcript`)
  return response.data
}

/**
 * Search transcripts by text
 * GET /api/v1/search/transcript
 */
export async function searchTranscripts(
  params: TranscriptSearchRequest
): Promise<TranscriptSearchResponse> {
  const response = await v1ApiClient.get<TranscriptSearchResponse>('/search/transcript', { params })
  return response.data
}

/**
 * Update a transcript segment
 * PATCH /api/v1/transcript-segments/:id
 */
export async function updateTranscriptSegment(
  id: number,
  data: { text_content: string }
): Promise<SegmentResponse> {
  const response = await v1ApiClient.patch<SegmentResponse>(`/transcript-segments/${id}`, data)
  return response.data
}

/**
 * Submit video transcript review
 * POST /api/v1/videos/:videoId/reviews
 */
export async function submitVideoReview(
  videoId: string,
  request?: SubmitReviewRequest
): Promise<VideoTranscriptReviewResponse> {
  const response = await v1ApiClient.post<VideoTranscriptReviewResponse>(
    `/videos/${videoId}/reviews`,
    request || {}
  )
  return response.data
}

/**
 * Get user's review status for a video
 * GET /api/v1/videos/:videoId/reviews/status
 */
export async function getUserReviewStatus(
  videoId: string
): Promise<{ video_id: string; has_reviewed: boolean }> {
  const response = await v1ApiClient.get<{ video_id: string; has_reviewed: boolean }>(
    `/videos/${videoId}/reviews/status`
  )
  return response.data
}

/**
 * Get video review statistics
 * GET /api/v1/videos/:videoId/reviews/stats
 */
export async function getVideoReviewStats(
  videoId: string
): Promise<{ video_id: string; review_count: number }> {
  const response = await v1ApiClient.get<{ video_id: string; review_count: number }>(
    `/videos/${videoId}/reviews/stats`
  )
  return response.data
}

// Export all functions as an object
export const videoApi = {
  fetchVideos,
  fetchVideoById,
  fetchVideoTranscript,
  searchTranscripts,
  updateTranscriptSegment,
  submitVideoReview,
  getUserReviewStatus,
  getVideoReviewStats,
}

export default videoApi
