import { api } from '@/api/client'
import { useOrganizationStore } from '@/stores/organizationStore'
import { useUserStore } from '@/stores/userStore'
import { useQuery } from '@tanstack/react-query'
import { AxiosError } from 'axios'
import { useEffect } from 'react'
import { type Organization } from '../types'

export const useFetchOrganization = () => {
  const setOrganization = useOrganizationStore((s) => s.setOrganization)
  const setOrganizationDomain = useOrganizationStore(
    (s) => s.setOrganizationDomain,
  )
  const setError = useUserStore((s) => s.setError)

  const query = useQuery<Organization, AxiosError>({
    queryKey: ['organization'],
    queryFn: async (): Promise<Organization> => {
      const { data } = await api.get('/api/v1/organization')
      return data
      // return mockOrganization
    },
    retry: true,
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 10000),
    refetchOnWindowFocus: false,
  })

  useEffect(() => {
    if (query.data) {
      setOrganization(query.data)
      setOrganizationDomain(query.data.title)
    }
  }, [query.data, setOrganization, setOrganizationDomain])

  useEffect(() => {
    if (query.error) {
      setError(query.error.message ?? 'Failed to fetch organization')
    }
  }, [query.error, setError])

  return query
}
