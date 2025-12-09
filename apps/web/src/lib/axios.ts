import axios from 'axios'

const API_URL = import.meta.env.VITE_API_URL + (import.meta.env.VITE_API_TAG || '/api/v1')

/**
 * Axios instance for API calls
 * Configured with base URL and credentials support
 */
const axiosInstance = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Send cookies with requests
})

export default axiosInstance
