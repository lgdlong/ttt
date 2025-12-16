import axios, { type AxiosError, type AxiosInstance } from 'axios'
import type { ErrorResponse } from '~/types/video'

const API_BASE_URL = import.meta.env.VITE_API_URL || ''

/**
 * Creates a new Axios instance for a specific API version.
 * This factory ensures all API clients share the same base configuration
 * for interceptors, credentials, and error handling.
 *
 * @param version - The API version ('v1', 'v2', etc.)
 * @returns A pre-configured Axios instance.
 */
export const createApiClient = (version: 'v1' | 'v2'): AxiosInstance => {
  const instance = axios.create({
    baseURL: `${API_BASE_URL}/api/${version}`,
    headers: {
      'Content-Type': 'application/json',
    },
    withCredentials: true, // Send cookies with all requests
  })

  // Apply a shared response interceptor for unified error handling
  instance.interceptors.response.use(
    (response) => response,
    (error: AxiosError<ErrorResponse>) => {
      // Extract a more meaningful error message from the response
      const errorMessage =
        error.response?.data?.message || error.response?.data?.error || error.message
      
      // We throw a new Error so that React Query and other data fetching libraries
      // can catch it and manage the error state properly.
      return Promise.reject(new Error(errorMessage))
    }
  )

  return instance
}

// Optionally, create and export pre-configured instances for common versions
export const v1ApiClient = createApiClient('v1')
export const v2ApiClient = createApiClient('v2')
