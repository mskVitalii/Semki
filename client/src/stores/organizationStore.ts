import { type Organization } from '@/common/types'
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
}

export const useOrganizationStore = create<OrganizationState>()(
  persist(
    (set) => ({
      organization: null,
      organizationDomain: getOrgDomainFromHref(),
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
    }),
    {
      name: 'organization-storage',
    },
  ),
)
function getOrgDomainFromHref(): string {
  const hostname = window.location.hostname
  const isLocalDev = hostname === 'localhost' || hostname === '127.0.0.1'
  const organization_domain = isLocalDev
    ? 'mockOrganization'
    : hostname.split('.')[0]

  return organization_domain.toLocaleLowerCase().trim().replaceAll(' ', '-')
}
