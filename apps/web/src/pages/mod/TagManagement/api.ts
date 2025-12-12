import axiosInstance from '~/lib/axios'
import type {
  TagResponse,
  TagListResponse,
  CreateTagRequest,
  MergeTagsRequest,
  MergeTagsResponse,
  UpdateTagApprovalRequest,
} from '~/types/tag'

// V2 API base path
const API_BASE = '/v2/mod/tags'

// ============================================================
// Tag List & Search
// ============================================================

interface FetchTagsParams {
  page: number
  pageSize: number
  query?: string
}

interface FetchTagsResult {
  tags: TagResponse[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export const fetchTags = async (params: FetchTagsParams): Promise<FetchTagsResult> => {
  const response = await axiosInstance.get(API_BASE, {
    params: {
      page: params.page,
      limit: params.pageSize,
      query: params.query || undefined,
    },
  })

  // V2 API returns { status, data, metadata: { pagination } }
  const { data, metadata } = response.data
  const pagination = metadata?.pagination || {}

  return {
    tags: data || [],
    total: pagination.total || 0,
    page: pagination.page || 1,
    pageSize: pagination.limit || 20,
    totalPages: pagination.total_pages || 1,
  }
}

export const searchTags = async (query: string, limit = 10): Promise<TagResponse[]> => {
  const response = await axiosInstance.get(`${API_BASE}/search`, {
    params: { q: query, limit },
  })
  // V2 API returns { status, data }
  return response.data.data || []
}

// ============================================================
// Tag CRUD
// ============================================================

export const createTag = async (data: CreateTagRequest): Promise<TagResponse> => {
  const response = await axiosInstance.post(API_BASE, data)
  // V2 API returns { status, data } with TagResolveResponse format
  const tagData = response.data.data
  return {
    id: tagData.id,
    name: tagData.name,
    is_approved: false, // New tags are not approved by default
  }
}

export const getTag = async (id: string): Promise<TagResponse> => {
  const response = await axiosInstance.get(`${API_BASE}/${id}`)
  return response.data.data
}

// ============================================================
// Tag Merge (Manual)
// ============================================================

export const mergeTags = async (data: MergeTagsRequest): Promise<MergeTagsResponse> => {
  const response = await axiosInstance.post(`${API_BASE}/merge`, data)
  return response.data.data
}

// ============================================================
// Tag Approval
// ============================================================

export const updateTagApproval = async ({
  id,
  data,
}: {
  id: string
  data: UpdateTagApprovalRequest
}): Promise<TagResponse> => {
  const response = await axiosInstance.patch(`${API_BASE}/${id}/approve`, data)
  return response.data.data
}

export type { FetchTagsResult, TagListResponse }
