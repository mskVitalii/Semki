// stores/authStore.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthStore {
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
      setAuth: (accessToken, refreshToken) =>
        set({ accessToken, refreshToken }),
      logout: () => set({ accessToken: null, refreshToken: null }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
    },
  ),
)
