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
export const deleteVideo = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/v1/mod/videos/${id}`)
}

/**
 * Add tags to a video
 * POST /api/v1/mod/videos/:videoId/tags
 * V1 endpoint (legacy)
 */
export const addTagsToVideo = async (videoId: number, tagIds: number[]): Promise<Video> => {
  const response = await axiosInstance.post(`/v1/mod/videos/${videoId}/tags`, {
    tag_ids: tagIds,
  })
  return response.data
}

/**
 * Remove a tag from a video
 * DELETE /api/v1/mod/videos/:videoId/tags/:tagId
 * V1 endpoint (legacy) - Note: accepts string tagId for v2 compatibility
 */
export const removeTagFromVideo = async (videoId: number, tagId: string): Promise<void> => {
  await axiosInstance.delete(`/v1/mod/videos/${videoId}/tags/${tagId}`)
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
  createVideo,
  deleteVideo,
  addTagsToVideo,
  removeTagFromVideo,
  fetchVideoPreview,
}

export default modApi
