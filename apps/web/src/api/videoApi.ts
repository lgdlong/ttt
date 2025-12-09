import axios, { type AxiosError } from 'axios'
import type {
  ListVideoRequest,
  VideoListResponse,
  VideoDetailResponse,
  TranscriptResponse,
  TranscriptSearchRequest,
  TranscriptSearchResponse,
  TagSearchRequest,
  TagSearchResponse,
  TagResponse,
  ErrorResponse,
} from '~/types/video'

const API_URL = import.meta.env.VITE_API_URL + (import.meta.env.VITE_API_TAG || '/api/v1')

/**
 * Video API Service
 * Handles all video-related API calls to backend
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
 * Search tags by semantic similarity
 * GET /api/search/tags
 */
export async function searchTags(params: TagSearchRequest): Promise<TagSearchResponse> {
  const response = await apiClient.get<TagSearchResponse>('/search/tags', { params })
  return response.data
}

/**
 * Get all unique tags (for filter/category list)
 * Note: This is derived from searchTags with empty query or needs a new endpoint
 */
export async function fetchAllTags(): Promise<TagResponse[]> {
  // TODO: Backend should provide a dedicated endpoint like GET /api/tags
  // For now, we return empty array - frontend will need to handle this
  return []
}

// Export all functions as an object
export const videoApi = {
  fetchVideos,
  fetchVideoById,
  fetchVideoTranscript,
  searchTranscripts,
  searchTags,
  fetchAllTags,
}

export default videoApi
