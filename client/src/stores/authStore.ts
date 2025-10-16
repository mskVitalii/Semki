// stores/authStore.ts
import { getUserClaims, type UserClaims } from '@/common/jwt'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthStore {
  claims: UserClaims | null
  accessToken: string | null
  refreshToken: string | null
  setAuth: (accessToken: string, refreshToken: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      claims: null,
      setAuth: (accessToken, refreshToken) =>
        set(() => {
          const claims = getUserClaims(accessToken)
          console.log('setAuth', claims, accessToken, refreshToken)
          return { accessToken, refreshToken, claims }
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
      }),
    },
  ),
)
