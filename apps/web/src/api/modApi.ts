import axiosInstance from '~/lib/axios'
import type { Video, VideoListRequest, ModVideoListResponse } from '~/types/video'
import type { TagResponse } from '~/types/tag'

/**
 * Moderator API Service
 * Handles moderator-specific video and tag management operations
 *
 * Version: Mixed - V1 for videos, V2 for canonical tags
 */

/**
 * Fetch videos for moderator dashboard
 * GET /api/v1/mod/videos
 * V1 endpoint (legacy)
 */
export const fetchModVideos = async (
  params: VideoListRequest & { has_transcript?: string }
): Promise<ModVideoListResponse> => {
  const response = await axiosInstance.get('/v1/mod/videos', {
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
 * V2 endpoint (new canonical tag system)
 */
export const fetchAllTags = async (): Promise<{ tags: TagResponse[] }> => {
  const response = await axiosInstance.get('/v2/mod/tags', {
    params: { limit: 1000 },
  })
  // v2 API returns { status, data, metadata }
  return { tags: response.data.data || [] }
}

/**
 * Search approved canonical tags
 * GET /api/v2/mod/tags/search
 * V2 endpoint - only returns approved tags
 */
export const searchApprovedTags = async (query: string, limit = 20): Promise<TagResponse[]> => {
  const response = await axiosInstance.get('/v2/mod/tags/search', {
    params: { q: query, limit, approved_only: true },
  })
  // v2 API returns { status, data }
  return response.data.data || []
}

/**
 * Get video's canonical tags
 * GET /api/v2/mod/videos/:videoId/tags
 * V2 endpoint
 */
export const getVideoTags = async (videoId: string): Promise<TagResponse[]> => {
  const response = await axiosInstance.get(`/v2/mod/videos/${videoId}/tags`)
  return response.data.data || []
}

/**
 * Create a new video
 * POST /api/v1/mod/videos
 * V1 endpoint (legacy)
 */
export const createVideo = async (youtubeId: string, tagIds?: number[]): Promise<Video> => {
  const response = await axiosInstance.post('/v1/mod/videos', {
    youtube_id: youtubeId,
    tag_ids: tagIds,
  })
  return response.data
}

/**
 * Delete a video
 * DELETE /api/v1/mod/videos/:id
 * V1 endpoint (legacy)
 */
export const deleteVideo = async (id: string): Promise<void> => {
  await axiosInstance.delete(`/v1/mod/videos/${id}`)
}

/**
 * Add tags to a video (V2 - one tag at a time)
 * POST /api/v2/mod/videos/:videoId/tags
 * Calls backend for each tag_id due to API limitation (1 tag per request)
 * Returns updated tags list
 */
export const addTagsToVideo = async (videoId: string, tagIds: string[]): Promise<TagResponse[]> => {
  // V2 API only accepts one tag per request, so loop through tagIds
  for (const tagId of tagIds) {
    await axiosInstance.post(`/v2/mod/videos/${videoId}/tags`, {
      tag_id: tagId,
    })
  }
  // Fetch updated tags for the video
  const response = await axiosInstance.get(`/v2/mod/videos/${videoId}/tags`)
  return response.data.data || []
}

/**
 * Remove a tag from a video
 * DELETE /api/v2/mod/videos/:videoId/tags/:tagId
 * V2 endpoint
 */
export const removeTagFromVideo = async (videoId: string, tagId: string): Promise<void> => {
  await axiosInstance.delete(`/v2/mod/videos/${videoId}/tags/${tagId}`)
}

/**
 * Fetch video preview by YouTube ID
 * GET /api/v1/mod/videos/preview/:youtubeId
 * V1 endpoint (legacy)
 */
export const fetchVideoPreview = async (youtubeId: string): Promise<Video> => {
  const response = await axiosInstance.get(`/v1/mod/videos/preview/${youtubeId}`)
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
