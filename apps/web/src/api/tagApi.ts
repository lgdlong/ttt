import axios, { type AxiosError } from 'axios'
import type { TagSearchRequest, TagSearchResponse, TagResponse, ErrorResponse } from '~/types/video'

// Tag API uses v1 endpoints (legacy)
const API_URL = import.meta.env.VITE_API_URL + (import.meta.env.VITE_API_TAG || '/api') + '/v1'

/**
 * Tag API Service
 * Handles all tag-related API calls to backend
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

// Tag list response structure from backend
interface TagListApiResponse {
  success: boolean
  message: string
  data: TagResponse[]
  metadata?: {
    pagination?: {
      page: number
      limit: number
      total_items: number
      total_pages: number
    }
  }
}

// Single tag response structure from backend
interface SingleTagApiResponse {
  success: boolean
  message: string
  data: TagResponse
}

/**
 * Search tags by semantic similarity
 * GET /api/v1/search/tags
 */
export async function searchTags(params: TagSearchRequest): Promise<TagSearchResponse> {
  const response = await apiClient.get<TagSearchResponse>('/search/tags', { params })
  return response.data
}

/**
 * Get all approved tags with pagination
 * GET /api/v1/tags
 * Only shows tags that have been approved by moderators
 */
export async function fetchAllTagsApproved(params?: {
  page?: number
  limit?: number
}): Promise<TagResponse[]> {
  const response = await apiClient.get<TagListApiResponse>('/tags', {
    params: { limit: 100, ...params }, // Default to 100 tags
  })
  return response.data.data || []
}

/**
 * Fetch single tag by ID
 * GET /api/v1/tags/:id
 */
export async function fetchTagById(id: string): Promise<TagResponse> {
  const response = await apiClient.get<SingleTagApiResponse>(`/tags/${id}`)
  return response.data.data
}

// Export all functions as an object
export const tagApi = {
  searchTags,
  fetchAllTagsApproved,
  fetchTagById,
}

export default tagApi
