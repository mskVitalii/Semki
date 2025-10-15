import { mockOrganization, type Organization } from '@/utils/types'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface OrganizationState {
  organization: Organization | null
  organizationDomain: string
  isLoading: boolean
  error: string | null
  isInitialized: boolean
  setOrganization: (org: Organization) => void
  setOrganizationDomain: (domain: string) => void
  setError: (error: string) => void
  fetchOrganization: () => Promise<void>
}

export const useOrganizationStore = create<OrganizationState>()(
  persist(
    (set, get) => ({
      organization: null,
      organizationDomain: '',
      isLoading: false,
      error: null,
      isInitialized: false,

      setOrganization: (org) => set({ organization: org, isLoading: false }),

      setOrganizationDomain: (domain) =>
        set({
          organizationDomain: domain
            .toLocaleLowerCase()
            .trim()
            .replaceAll(' ', '-'),
        }),

      setError: (error) => set({ error, isLoading: false }),

      fetchOrganization: async () => {
        const hostname = window.location.hostname
        const isLocalDev = hostname === 'localhost' || hostname === '127.0.0.1'

        const organizationDomain = isLocalDev
          ? mockOrganization.title
              .toLocaleLowerCase()
              .trim()
              .replaceAll(' ', '-')
          : hostname.split('.')[0]

        if (isLocalDev) {
          console.log('Set isLocalDev', get().error)
          set({
            organizationDomain: organizationDomain,
            organization: mockOrganization,
          })
          return
        }
        console.log('After isLocalDev')

        const currentDomain = get().organizationDomain

        if (currentDomain && currentDomain !== organizationDomain) {
          // brand new start
          set({
            organization: null,
            organizationDomain,
            isLoading: true,
            error: null,
            isInitialized: false,
          })
        }
        if (get().isInitialized && currentDomain === organizationDomain) return

        set({
          organizationDomain: organizationDomain,
          isLoading: true,
          error: null,
        })

        try {
          const response = await fetch(
            `/api/organizations/${organizationDomain}`,
          )

          if (!response.ok) {
            throw new Error('Organization not found')
          }

          const organization: Organization = await response.json()
          set({ organization, isLoading: false })
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Unknown error',
            isLoading: false,
          })
        }
      },
    }),
    {
      name: 'organization-storage',
    },
  ),
)

export const initializeOrganization = () => {
  useOrganizationStore.getState().fetchOrganization()
}
