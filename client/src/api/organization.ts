import type { User } from '@/common/types'
import { api } from './client'

export type FetchUsersResponse = {
  users: User[]
  totalCount: number
}

export const fetchOrganizationUsers = async (
  page = 1,
  limit = 5,
  search?: string,
): Promise<FetchUsersResponse> => {
  const { data } = await api.get<FetchUsersResponse>(
    '/api/v1/organization/users',
    {
      params: { page, limit, search },
    },
  )
  return data
}
