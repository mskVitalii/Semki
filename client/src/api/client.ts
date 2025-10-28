// api/client.ts
import { useAuthStore } from '@/stores/authStore'
import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
})

// Request interceptor - adds token to each request
api.interceptors.request.use(
  (config) => {
    const accessToken = useAuthStore.getState().accessToken
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// Response interceptor - handle 401 by refresh_token
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      const refreshToken = useAuthStore.getState().refreshToken

      if (!refreshToken) {
        useAuthStore.getState().logout()
        if (!window.location.href.endsWith('/login')) {
          console.log('No refresh token, redirect to login')
          window.location.href = '/login'
        }
        return Promise.reject(error)
      }

      try {
        console.log('Try to refresh token')
        const result = await api.post('/api/v1/refresh_token', {
          refresh_token: refreshToken,
        })

        if (result.status !== 200) {
          throw new Error('Failed to refresh token')
        }

        const data = result.data
        useAuthStore.getState().setAuth(data.access_token, data.refresh_token)
        originalRequest.headers.Authorization = `Bearer ${data.access_token}`

        return api(originalRequest)
      } catch (refreshError) {
        useAuthStore.getState().logout()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  },
)
