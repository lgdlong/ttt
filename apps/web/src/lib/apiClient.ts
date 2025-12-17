import axios, { type AxiosError, type AxiosInstance } from 'axios'
import type { ErrorResponse } from '~/types/video'
import { refreshToken } from '~/api/authApi'
import { createEventBus } from '~/lib/eventBus'

const API_BASE_URL = import.meta.env.VITE_API_URL || ''

// Centralized event bus for cross-module communication
const eventBus = createEventBus()

// Function to create a new Axios instance
export const createApiClient = (version: 'v1' | 'v2'): AxiosInstance => {
  const instance = axios.create({
    baseURL: `${API_BASE_URL}/api/${version}`,
    headers: {
      'Content-Type': 'application/json',
    },
    withCredentials: true,
  })

  instance.interceptors.response.use(
    (response) => response,
    async (error: AxiosError<ErrorResponse>) => {
      const originalRequest = error.config as any; // Use 'as any' to add custom property
      if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
        const isLogout = originalRequest.url?.endsWith('/auth/logout');
        const isRefresh = originalRequest.url?.endsWith('/auth/refresh');

        if (isLogout || isRefresh) {
            eventBus.emit('auth:logout');
            return Promise.reject(error);
        }

        originalRequest._retry = true;
        try {
          await refreshToken();
          return instance(originalRequest);
        } catch (refreshError) {
          eventBus.emit('auth:logout');
          return Promise.reject(refreshError);
        }
      }
      const errorMessage =
        error.response?.data?.message || error.response?.data?.error || error.message;
      return Promise.reject(new Error(errorMessage));
    }
  );

  return instance
}

export const v1ApiClient = createApiClient('v1')
export const v2ApiClient = createApiClient('v2')
export { eventBus }