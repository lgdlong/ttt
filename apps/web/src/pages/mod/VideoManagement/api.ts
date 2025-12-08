import axiosInstance from '~/lib/axios'
import type { Video, VideoListRequest, ModVideoListResponse } from '~/types/video'
import type { TagResponse } from '~/types/tag'

export const fetchVideos = async (
  params: VideoListRequest & { has_transcript?: string }
): Promise<ModVideoListResponse> => {
  const response = await axiosInstance.get('/mod/videos', {
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

export const fetchAllTags = async (): Promise<{ tags: TagResponse[] }> => {
  const response = await axiosInstance.get('/mod/tags', {
    params: { page_size: 1000 },
  })
  return response.data
}

export const createVideo = async (youtubeId: string, tagIds?: number[]): Promise<Video> => {
  const response = await axiosInstance.post('/mod/videos', {
    youtube_id: youtubeId,
    tag_ids: tagIds,
  })
  return response.data
}

export const deleteVideo = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/mod/videos/${id}`)
}

export const addTagsToVideo = async (videoId: number, tagIds: number[]): Promise<Video> => {
  const response = await axiosInstance.post(`/mod/videos/${videoId}/tags`, {
    tag_ids: tagIds,
  })
  return response.data
}

export const removeTagFromVideo = async (videoId: number, tagId: number): Promise<void> => {
  await axiosInstance.delete(`/mod/videos/${videoId}/tags/${tagId}`)
}

export const fetchVideoPreview = async (youtubeId: string): Promise<Video> => {
  const response = await axiosInstance.get(`/mod/videos/preview/${youtubeId}`)
  return response.data
}
