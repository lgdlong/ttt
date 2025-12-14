import axios, { type AxiosError } from 'axios'
import type {
  ListVideoRequest,
  VideoListResponse,
  VideoDetailResponse,
  TranscriptResponse,
  TranscriptSearchRequest,
  TranscriptSearchResponse,
  ErrorResponse,
  SegmentResponse,
  SubmitReviewRequest,
  VideoTranscriptReviewResponse,
} from '~/types/video'

// Video API uses v1 endpoints (legacy)
const API_URL = import.meta.env.VITE_API_URL + (import.meta.env.VITE_API_TAG || '/api') + '/v1'

/**
 * Video API Service
 * Handles all video-related API calls to backend
 *
 * Version: V1 (legacy)
 */

// Create axios instance with default config
const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Response interceptor to handle errors
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ErrorResponse>) => {
    const errorMessage =
      error.response?.data?.message || error.response?.data?.error || error.message
    throw new Error(errorMessage)
  }
)

/**
 * Fetch videos with pagination and filters
 * GET /api/videos
 */
export async function fetchVideos(params?: ListVideoRequest): Promise<VideoListResponse> {
  const response = await apiClient.get<VideoListResponse>('/videos', { params })
  return response.data
}

/**
 * Fetch single video detail by ID
 * GET /api/videos/:id
 */
export async function fetchVideoById(id: string): Promise<VideoDetailResponse> {
  const response = await apiClient.get<VideoDetailResponse>(`/videos/${id}`)
  return response.data
}

/**
 * Fetch video transcript by video ID
 * GET /api/videos/:id/transcript
 */
export async function fetchVideoTranscript(videoId: string): Promise<TranscriptResponse> {
  const response = await apiClient.get<TranscriptResponse>(`/videos/${videoId}/transcript`)
  return response.data
}

/**
 * Search transcripts by text
 * GET /api/search/transcript
 */
export async function searchTranscripts(
  params: TranscriptSearchRequest
): Promise<TranscriptSearchResponse> {
  const response = await apiClient.get<TranscriptSearchResponse>('/search/transcript', { params })
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
  const response = await apiClient.patch<SegmentResponse>(`/transcript-segments/${id}`, data)
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
  const response = await apiClient.post<VideoTranscriptReviewResponse>(
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
  const response = await apiClient.get<{ video_id: string; has_reviewed: boolean }>(
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
  const response = await apiClient.get<{ video_id: string; review_count: number }>(
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
