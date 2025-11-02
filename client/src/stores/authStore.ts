// stores/authStore.ts
import { getUserClaims, type UserClaims } from '@/common/jwt'
import { OrganizationRoles } from '@/common/types'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthStore {
  claims: UserClaims | null
  accessToken: string | null
  refreshToken: string | null
  isAdmin: boolean | null
  setAuth: (accessToken: string, refreshToken: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      isAdmin: null,
      claims: null,
      setAuth: (accessToken, refreshToken) =>
        set(() => {
          const claims = getUserClaims(accessToken)
          const isAdmin = [
            OrganizationRoles.ADMIN,
            OrganizationRoles.OWNER,
          ].includes(claims?.organizationRole ?? OrganizationRoles.USER)
          return { accessToken, refreshToken, claims, isAdmin }
        }),
      logout: () =>
        set({ accessToken: null, refreshToken: null, claims: null }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        claims: state.claims,
        isAdmin: state.isAdmin,
      }),
    },
  ),
)
