import type { User } from '@/common/types'
import { api } from './client'

export interface InviteUserData {
  email: string
  name: string
  organizationRole?: string
  semantic?: {
    team?: string
    level?: string
    location?: string
    description?: string
  }
}

export const inviteUser = async (userData: InviteUserData): Promise<User> => {
  const { data } = await api.post<User>('/api/v1/user/invite', userData)
  return data
}

export const restoreUserAccount = async (userId: string) => {
  const { data } = await api.post(`/api/v1/users/${userId}/restore`)
  return data
}

export const deleteUserAccount = async (userId: string) => {
  const { data } = await api.delete(`/api/v1/users/${userId}`)
  return data
}
