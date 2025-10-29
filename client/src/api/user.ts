import type { User, UserStatus } from '@/common/types'
import { api } from './client'

export interface InviteUserData {
  email: string
  name: string
  organizationRole?: string
  semantic?: {
    team?: string
    level?: string
    location?: string
  }
}

export const inviteUser = async (userData: InviteUserData): Promise<User> => {
  const { data } = await api.post<User>('/api/v1/user/invite', userData)
  return data
}

export const updateUserStatus = async (userId: string, status: UserStatus) => {
  const { data } = await api.patch(`/api/v1/users/${userId}`, { status })
  return data
}
