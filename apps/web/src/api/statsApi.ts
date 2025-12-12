import axiosInstance from '~/lib/axios'

/**
 * Stats API Types
 */
export interface AdminStatsResponse {
  total_users: number
  active_users: number
  total_videos: number
  total_tags: number
}

export interface ModStatsResponse {
  total_videos: number
  total_tags: number
  videos_with_transcript: number
  videos_added_today: number
}

/**
 * Get admin dashboard statistics
 * GET /api/v1/admin/stats (v1 - legacy)
 * Requires admin role
 */
export async function getAdminStats(): Promise<AdminStatsResponse> {
  const response = await axiosInstance.get<AdminStatsResponse>('/v1/admin/stats')
  return response.data
}

/**
 * Get moderator dashboard statistics
 * GET /api/v1/mod/stats (v1 - legacy)
 * Requires mod or admin role
 */
export async function getModStats(): Promise<ModStatsResponse> {
  const response = await axiosInstance.get<ModStatsResponse>('/v1/mod/stats')
  return response.data
}
