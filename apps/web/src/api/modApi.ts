import { v1ApiClient, v2ApiClient } from '~/lib/apiClient'
import type { Video, VideoListRequest, ModVideoListResponse } from '~/types/video'
import type { TagResponse } from '~/types/tag'

/**
 * Moderator API Service
 * Handles moderator-specific video and tag management operations
 *
 * This service uses both v1 and v2 API clients to interact with
 * legacy and new endpoints respectively.
 */

/**
 * Fetch videos for moderator dashboard
 * GET /api/v1/mod/videos
 */
export const fetchModVideos = async (
  params: VideoListRequest & { has_transcript?: string }
): Promise<ModVideoListResponse> => {
  const response = await v1ApiClient.get('/mod/videos', {
    params: {
      page: params.page,
      page_size: params.page_size,
      q: params.q || undefined,
      tag_ids: params.tag_ids?.join(',') || undefined,
      has_transcript: params.has_transcript || undefined,
    },
  })
  return response.data
}

/**
 * Fetch all canonical tags
 * GET /api/v2/mod/tags
 */
export const fetchAllTags = async (): Promise<{ tags: TagResponse[] }> => {
  const response = await v2ApiClient.get('/mod/tags', {
    params: { limit: 1000 },
  })
  // v2 API returns { status, data, metadata }
  return { tags: response.data.data || [] }
}

/**
 * Search approved canonical tags
 * GET /api/v2/mod/tags/search
 */
export const searchApprovedTags = async (query: string, limit = 20): Promise<TagResponse[]> => {
  const response = await v2ApiClient.get('/mod/tags/search', {
    params: { q: query, limit, approved_only: true },
  })
  // v2 API returns { status, data }
  return response.data.data || []
}

/**
 * Get video's canonical tags
 * GET /api/v2/mod/videos/:videoId/tags
 */
export const getVideoTags = async (videoId: string): Promise<TagResponse[]> => {
  const response = await v2ApiClient.get(`/mod/videos/${videoId}/tags`)
  return response.data.data || []
}

/**
 * Create a new video
 * POST /api/v1/mod/videos
 */
export const createVideo = async (youtubeId: string, tagIds?: number[]): Promise<Video> => {
  const response = await v1ApiClient.post('/mod/videos', {
    youtube_id: youtubeId,
    tag_ids: tagIds,
  })
  return response.data
}

/**
 * Delete a video
 * DELETE /api/v1/mod/videos/:id
 */
export const deleteVideo = async (id: string): Promise<void> => {
  await v1ApiClient.delete(`/mod/videos/${id}`)
}

/**
 * Add tags to a video (V2 - one tag at a time)
 * POST /api/v2/mod/videos/:videoId/tags
 */
export const addTagsToVideo = async (videoId: string, tagIds: string[]): Promise<TagResponse[]> => {
  for (const tagId of tagIds) {
    await v2ApiClient.post(`/mod/videos/${videoId}/tags`, {
      tag_id: tagId,
    })
  }
  const response = await v2ApiClient.get(`/mod/videos/${videoId}/tags`)
  return response.data.data || []
}

/**
 * Remove a tag from a video
 * DELETE /api/v2/mod/videos/:videoId/tags/:tagId
 */
export const removeTagFromVideo = async (videoId: string, tagId: string): Promise<void> => {
  await v2ApiClient.delete(`/mod/videos/${videoId}/tags/${tagId}`)
}

/**
 * Fetch video preview by YouTube ID
 * GET /api/v1/mod/videos/preview/:youtubeId
 */
export const fetchVideoPreview = async (youtubeId: string): Promise<Video> => {
  const response = await v1ApiClient.get(`/mod/videos/preview/${youtubeId}`)
  return response.data
}

// Export all functions as an object for convenience
export const modApi = {
  fetchModVideos,
  fetchAllTags,
  searchApprovedTags,
  getVideoTags,
  createVideo,
  deleteVideo,
  addTagsToVideo,
  removeTagFromVideo,
  fetchVideoPreview,
}

export default modApi
