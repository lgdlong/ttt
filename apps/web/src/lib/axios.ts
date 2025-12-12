import axios from 'axios'

// Base API configuration - now defaults to /api (which maps to v1 endpoints)
// For v2 endpoints, use explicit /v2 prefix in request paths
const API_BASE = import.meta.env.VITE_API_URL
const API_TAG = import.meta.env.VITE_API_TAG || '/api'
const API_URL = API_BASE + API_TAG

/**
 * Axios instance for API calls
 * Configured with base URL and credentials support
 *
 * Endpoint versioning strategy:
 * - Default (v1): /mod/videos, /mod/tags (legacy), /admin/stats, /auth, etc.
 * - V2 endpoints: Use explicit /v2 prefix in request paths (e.g., /v2/mod/tags, /v2/mod/videos/{id}/tags)
 */
const axiosInstance = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Send cookies with requests
})

export default axiosInstance
export { API_BASE, API_TAG, API_URL }
