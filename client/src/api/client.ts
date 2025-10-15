// api/client.ts
import { useAuthStore } from '@/stores/authStore'
import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
})

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true // защита от бесконечного цикла

      try {
        const refreshToken = useAuthStore.getState().refreshToken
        const { data } = await axios.post('/auth/refresh', { refreshToken })
        useAuthStore.getState().setAuth(data.access, data.refresh)

        originalRequest.headers.Authorization = `Bearer ${data.access}`
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
