import { v1ApiClient } from '~/lib/apiClient'
import type { UserResponse, LoginRequest, SignupRequest } from '~/types/user'

/**
 * Auth API Service
 * Handles authentication-related API calls to the v1 backend
 */

export interface UpdateMePayload {
  full_name?: string
  email?: string
}

/**
 * Fetch the currently authenticated user's profile
 * GET /api/v1/auth/me
 */
export const getMe = async (): Promise<UserResponse> => {
  const response = await v1ApiClient.get<UserResponse>('/auth/me')
  return response.data
}

/**
 * Update the currently authenticated user's profile
 * PATCH /api/v1/auth/me
 */
export const updateMe = async (payload: UpdateMePayload): Promise<UserResponse> => {
  const response = await v1ApiClient.patch<UserResponse>('/auth/me', payload)
  return response.data
}

/**
 * Log in a user
 * POST /api/v1/auth/login
 */
export const login = async (payload: LoginRequest): Promise<{ user: UserResponse }> => {
  const response = await v1ApiClient.post<{ user: UserResponse }>('/auth/login', payload)
  return response.data
}

/**
 * Sign up a new user
 * POST /api/v1/auth/signup
 */
export const signup = async (payload: SignupRequest): Promise<{ user: UserResponse }> => {
  const response = await v1ApiClient.post<{ user: UserResponse }>('/auth/signup', payload)
  return response.data
}

/**
 * Log out the current user and revoke their refresh token
 * POST /api/v1/auth/logout
 */
export const logout = async (): Promise<void> => {
  await v1ApiClient.post('/auth/logout')
}

/**
 * Request a new access token using the refresh token (HttpOnly cookie)
 * POST /api/v1/auth/refresh-token
 */
export const refreshToken = async (): Promise<void> => {
  await v1ApiClient.post('/auth/refresh')
}



/**

 * Initiates Google OAuth flow by fetching the auth URL and redirecting.

 * GET /api/v1/auth/google

 */

export const loginWithGoogle = async (): Promise<void> => {

  const response = await v1ApiClient.get<{ url: string }>('/auth/google')

  const { url } = response.data

  // Redirect the browser to the Google auth page

  window.location.href = url

}
