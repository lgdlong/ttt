// V2 API Response - uses canonical-alias architecture
export interface TagResponse {
  id: string // UUID
  name: string
  is_approved: boolean
}

// V1 Legacy - kept for backward compatibility (DEPRECATED)
export interface LegacyTagResponse {
  id: number
  name: string
  description?: string
  video_count?: number
  created_at: string
  updated_at: string
}

export interface CreateTagRequest {
  name: string
}

export interface UpdateTagRequest {
  name?: string
  description?: string
}

export interface TagListResponse {
  data: TagResponse[]
  pagination: {
    page: number
    limit: number
    total: number
    total_pages: number
  }
}

export interface ListTagRequest {
  page?: number
  limit?: number
  query?: string
}

export interface AddTagsToVideoRequest {
  tag_id?: string // UUID
  tag_name?: string
}

// Merge tags request
export interface MergeTagsRequest {
  source_id: string // UUID - tag to be merged
  target_id: string // UUID - target canonical tag
}

export interface MergeTagsResponse {
  target_tag: TagResponse
  merged_alias_count: number
  source_tag_deleted: boolean
}

// Approval request
export interface UpdateTagApprovalRequest {
  is_approved: boolean
}
