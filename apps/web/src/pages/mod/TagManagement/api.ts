import axiosInstance from '~/lib/axios'
import type { TagResponse } from '~/types/tag'

interface TagListResponse {
  tags: TagResponse[]
  total: number
  page: number
  page_size: number
}

export const fetchTags = async (params: {
  page: number
  pageSize: number
  q?: string
}): Promise<TagListResponse> => {
  const response = await axiosInstance.get('/mod/tags', {
    params: {
      page: params.page,
      page_size: params.pageSize,
      q: params.q || undefined,
    },
  })
  return response.data
}

export interface CreateTagRequest {
  name: string
  description?: string
}

export interface UpdateTagRequest {
  name: string
  description?: string
}

export const createTag = async (data: CreateTagRequest): Promise<TagResponse> => {
  const response = await axiosInstance.post('/mod/tags', data)
  return response.data
}

export const updateTag = async ({
  id,
  data,
}: {
  id: number
  data: UpdateTagRequest
}): Promise<TagResponse> => {
  const response = await axiosInstance.put(`/mod/tags/${id}`, data)
  return response.data
}

export const deleteTag = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/mod/tags/${id}`)
}

export type { TagListResponse }
