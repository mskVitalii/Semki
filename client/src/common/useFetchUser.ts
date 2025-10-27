import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { useUserStore } from '@/stores/userStore'
import { useQuery } from '@tanstack/react-query'
import { AxiosError } from 'axios'
import { useEffect } from 'react'
import { type User } from './types'

export const useFetchUser = () => {
  const claims = useAuthStore((s) => s.claims)
  const setUser = useUserStore((s) => s.setUser)
  const setError = useUserStore((s) => s.setError)

  const query = useQuery<User | null, AxiosError>({
    queryKey: ['user'],
    queryFn: async () => {
      if (!claims || !claims._id) {
        return null
      }
      const { data } = await api.get(`/api/v1/user/${claims._id}`)
      return data
    },
    retry: true,
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 10000),
    refetchOnWindowFocus: false,
  })

  useEffect(() => {
    if (query.data) setUser(query.data)
  }, [query.data, setUser])

  useEffect(() => {
    if (query.error) setError(query.error.message ?? 'Failed to fetch user')
  }, [query.error, setError])

  return query
}
