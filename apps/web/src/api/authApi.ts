import axios, { type AxiosError } from 'axios'
import type {
  LoginRequest,
  SignupRequest,
  AuthResponse,
  UserResponse,
  CreateUserRequest,
  UpdateUserRequest,
  ListUserRequest,
  UserListResponse,
  GoogleAuthURLResponse,
  SessionResponse,
} from '~/types/user'
import type { ErrorResponse } from '~/types/video'

const API_URL = import.meta.env.VITE_API_URL + (import.meta.env.VITE_API_TAG || '/api/v1')

/**
 * Auth & User API Service
 * Handles authentication and user management API calls
 * Uses cookies for token storage (httpOnly cookies set by backend)
 */

// Create axios instance with default config
const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Send cookies with requests
})

// Response interceptor to handle errors and token expiration
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ErrorResponse>) => {
    // TODO: TEMPORARILY DISABLED - Session and Refresh Token
    // const originalRequest = error.config

    // Handle 401 Unauthorized - try to refresh token
    // if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
    //   originalRequest._retry = true

    //   try {
    //     // Attempt to refresh the token
    //     await apiClient.post('/auth/refresh')
    //     // Retry the original request
    //     return apiClient(originalRequest)
    //   } catch {
    //     // Refresh failed, clear local state and redirect to login
    //     localStorage.removeItem('user')
    //     window.location.href = '/login'
    //   }
    // }

    const errorMessage =
      error.response?.data?.message || error.response?.data?.error || error.message
    throw new Error(errorMessage)
  }
)

// Extend AxiosRequestConfig to include _retry property
declare module 'axios' {
  export interface InternalAxiosRequestConfig {
    _retry?: boolean
  }
}

// ============ Authentication APIs ============

/**
 * Login with username and password
 * POST /api/auth/login
 * Token is stored in httpOnly cookie by backend
 */
export async function login(req: LoginRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/login', req)
  // Store user in localStorage for quick access
  localStorage.setItem('user', JSON.stringify(response.data.user))
  return response.data
}

/**
 * Signup new user account
 * POST /api/auth/signup
 * Token is stored in httpOnly cookie by backend
 */
export async function signup(req: SignupRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/signup', req)
  // Store user in localStorage for quick access
  localStorage.setItem('user', JSON.stringify(response.data.user))
  return response.data
}

/**
 * Logout - clear session and cookies
 * POST /api/auth/logout
 */
export async function logout(): Promise<void> {
  try {
    await apiClient.post('/auth/logout')
  } catch {
    // Ignore errors on logout
  } finally {
    localStorage.removeItem('user')
    window.location.href = '/login'
  }
}

/**
 * Refresh access token using refresh token cookie
 * POST /api/auth/refresh
 * TODO: TEMPORARILY DISABLED - Session and Refresh Token
 */
/*
export async function refreshToken(): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/refresh')
  // Update user in localStorage
  localStorage.setItem('user', JSON.stringify(response.data.user))
  return response.data
}
*/

/**
 * Get current user info from server
 * GET /api/auth/me
 */
export async function getMe(): Promise<UserResponse> {
  const response = await apiClient.get<UserResponse>('/auth/me')
  // Update user in localStorage
  localStorage.setItem('user', JSON.stringify(response.data))
  return response.data
}

/**
 * Get current user from localStorage (sync)
 */
export function getCurrentUser(): UserResponse | null {
  const userStr = localStorage.getItem('user')
  if (!userStr) return null
  try {
    return JSON.parse(userStr) as UserResponse
  } catch {
    return null
  }
}

/**
 * Check if user is authenticated (has user data in localStorage)
 * Note: This is a quick check, actual auth is validated by server via cookies
 */
export function isAuthenticated(): boolean {
  return !!localStorage.getItem('user')
}

/**
 * Get active sessions
 * GET /api/auth/sessions
 */
export async function getActiveSessions(): Promise<SessionResponse[]> {
  const response = await apiClient.get<SessionResponse[]>('/auth/sessions')
  return response.data
}

// ============ Google OAuth APIs ============

/**
 * Get Google OAuth URL
 * GET /api/auth/google
 */
export async function getGoogleAuthURL(): Promise<string> {
  const response = await apiClient.get<GoogleAuthURLResponse>('/auth/google')
  return response.data.url
}

/**
 * Initiate Google OAuth login
 * Redirects to Google OAuth consent page
 */
export async function loginWithGoogle(): Promise<void> {
  const url = await getGoogleAuthURL()
  window.location.href = url
}

// ============ User Management APIs (Admin) ============

/**
 * Create a new user (admin only)
 * POST /api/users
 */
export async function createUser(req: CreateUserRequest): Promise<UserResponse> {
  const response = await apiClient.post<UserResponse>('/users', req)
  return response.data
}

/**
 * Get user by ID (admin only)
 * GET /api/users/:id
 */
export async function getUserById(id: string): Promise<UserResponse> {
  const response = await apiClient.get<UserResponse>(`/users/${id}`)
  return response.data
}

/**
 * Update user (admin only)
 * PUT /api/users/:id
 */
export async function updateUser(id: string, req: UpdateUserRequest): Promise<UserResponse> {
  const response = await apiClient.put<UserResponse>(`/users/${id}`, req)
  return response.data
}

/**
 * Delete user (soft delete, admin only)
 * DELETE /api/users/:id
 */
export async function deleteUser(id: string): Promise<void> {
  await apiClient.delete(`/users/${id}`)
}

/**
 * List users with pagination and filters (admin only)
 * GET /api/users
 */
export async function listUsers(params?: ListUserRequest): Promise<UserListResponse> {
  const response = await apiClient.get<UserListResponse>('/users', { params })
  return response.data
}

// ============ Helper Functions ============

/**
 * Get redirect path based on user role
 */
export function getRedirectPathByRole(role: string): string {
  switch (role) {
    case 'admin':
      return '/admin'
    case 'mod':
      return '/mod'
    default:
      return '/'
  }
}

/**
 * Check if user has required role
 */
export function hasRole(requiredRoles: string[]): boolean {
  const user = getCurrentUser()
  if (!user) return false
  return requiredRoles.includes(user.role)
}

/**
 * Check if user is admin
 */
export function isAdmin(): boolean {
  return hasRole(['admin'])
}

/**
 * Check if user is moderator or admin
 */
export function isModerator(): boolean {
  return hasRole(['admin', 'mod'])
}
