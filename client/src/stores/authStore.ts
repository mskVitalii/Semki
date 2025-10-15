// stores/authStore.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Organization } from '../utils/types'

interface AuthStore {
  accessToken: string | null
  refreshToken: string | null
  organization: Organization | null
  setAuth: (accessToken: string, refreshToken: string) => void
  setOrganization: (org: Organization) => void
  logout: () => void
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      organization: null,
      setAuth: (accessToken, refreshToken) =>
        set({ accessToken, refreshToken }),
      setOrganization: (org) => set({ organization: org }),
      logout: () => set({ accessToken: null, organization: null }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        organization: state.organization,
      }),
    },
  ),
)
