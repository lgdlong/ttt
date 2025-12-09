export interface TagResponse {
  id: number
  name: string
  description?: string
  video_count?: number
  created_at: string
  updated_at: string
}

export interface CreateTagRequest {
  name: string
  description?: string
}

export interface UpdateTagRequest {
  name?: string
  description?: string
}

export interface TagListResponse {
  tags: TagResponse[]
  total: number
  page: number
  page_size: number
}

export interface ListTagRequest {
  page?: number
  page_size?: number
  q?: string
}

export interface AddTagsToVideoRequest {
  tag_ids?: number[]
  tag_names?: string[]
}
