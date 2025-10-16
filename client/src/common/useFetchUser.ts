import { useUserStore } from '@/stores/userStore'
import { useQuery } from '@tanstack/react-query'
import { AxiosError } from 'axios'
import { useEffect } from 'react'
import { mockUser, type User } from './types'

export const useFetchUser = () => {
  const setUser = useUserStore((s) => s.setUser)
  const setError = useUserStore((s) => s.setError)

  const query = useQuery<User, AxiosError>({
    queryKey: ['user'],
    queryFn: async (): Promise<User> => {
      // const { data } = await api.get('/api/v1/user')
      // return data
      return mockUser
    },
  })

  useEffect(() => {
    if (query.data) {
      setUser(query.data)
    }
  }, [query.data, setUser])

  useEffect(() => {
    if (query.error) {
      setError(query.error.message ?? 'Failed to fetch user')
    }
  }, [query.error, setError])

  return query
}
