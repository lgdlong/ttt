import { v1ApiClient } from '~/lib/apiClient'
import type { UserResponse, ListUserRequest, UserListResponse, UpdateUserRequest } from '~/types/user'

/**
 * User Management API Service
 * Handles admin-only operations for managing users.
 */

/**
 * Fetch a paginated list of users
 * GET /api/v1/users
 */
export const listUsers = async (params: ListUserRequest): Promise<UserListResponse> => {
  const response = await v1ApiClient.get<UserListResponse>('/users', { params })
  return response.data
}

/**
 * Update a user's information by ID
 * PUT /api/v1/users/:id
 */
export const updateUser = async ({
  id,
  data,
}: {
  id: string
  data: UpdateUserRequest
}): Promise<UserResponse> => {
  const response = await v1ApiClient.put<UserResponse>(`/users/${id}`, data)
  return response.data
}
