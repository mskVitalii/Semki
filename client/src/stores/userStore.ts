import { OrganizationRoles, type User } from '@/common/types'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface UserState {
  user: User | null
  isAdmin: boolean
  isLoading: boolean
  error: string | null
  isInitialized: boolean
  setUser: (user: User) => void
  setError: (error: string) => void
}

export const useUserStore = create<UserState>()(
  persist(
    (set) => ({
      user: null,
      organizationDomain: '',
      isLoading: false,
      error: null,
      isInitialized: false,
      isAdmin: false,

      setUser: (user) =>
        set({
          user,
          isLoading: false,
          isAdmin:
            user.organizationRole === OrganizationRoles.OWNER ||
            user.organizationRole === OrganizationRoles.ADMIN,
        }),

      setError: (error) => set({ error, isLoading: false }),
    }),
    {
      name: 'user-storage',
    },
  ),
)
