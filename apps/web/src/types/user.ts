/**
 * User & Authentication Type Definitions
 * Matching backend DTOs from apps/api/internal/dto/user.go
 */

// ===== Enums =====

export type UserRole = 'user' | 'admin' | 'mod'

// ===== API Request Types =====

export interface LoginRequest {
  username: string
  password: string
}

export interface SignupRequest {
  username: string // min: 3, max: 50
  email: string
  password: string // min: 6
  full_name?: string // max: 100
}

export interface CreateUserRequest {
  username: string // min: 3, max: 50
  email: string
  password: string // min: 6
  full_name?: string // max: 100
  role?: UserRole
}

export interface UpdateUserRequest {
  username?: string // min: 3, max: 50
  email?: string
  password?: string // min: 6
  full_name?: string // max: 100
  role?: UserRole
  is_active?: boolean
}

export interface ListUserRequest {
  page?: number // Default: 1, min: 1
  limit?: number // Default: 20, min: 1, max: 100
  role?: UserRole
  is_active?: boolean
}

// ===== API Response Types =====

export interface UserResponse {
  id: string
  username: string
  email: string
  full_name: string
  role: UserRole
  is_active: boolean
  created_at: string // ISO timestamp
  updated_at: string // ISO timestamp
}

export interface AuthResponse {
  user: UserResponse
  token?: string // JWT token (optional when using cookies)
}

export interface GoogleAuthURLResponse {
  url: string
}

export interface SessionResponse {
  id: string
  user_agent: string
  client_ip: string
  created_at: string
  expires_at: string
  is_blocked: boolean
}

export interface SocialAccountResponse {
  id: string
  provider: string
  email: string
  created_at: string
}

export interface UserListResponse {
  data: UserResponse[]
  pagination: PaginationMetadata
}

// Re-export PaginationMetadata from video types
import type { PaginationMetadata } from './video'
export type { PaginationMetadata }
