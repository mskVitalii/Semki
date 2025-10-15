import { useOrganizationStore } from '@/stores/organizationStore'
import { useEffect } from 'react'
import { Navigate, Outlet } from 'react-router-dom'
import { mockOrganization } from './types'

export const OrganizationRoute = ({
  children,
}: {
  children?: React.ReactNode
}) => {
  const { setOrganizationDomain, isLoading, error } = useOrganizationStore()

  useEffect(() => {
    const hostname = window.location.hostname
    const isLocalDev = hostname === 'localhost' || hostname === '127.0.0.1'
    if (isLocalDev) {
      setOrganizationDomain(
        mockOrganization.title.toLocaleLowerCase().trim().replaceAll(' ', '-'),
      )
      return
    }
    const organizationDomain = hostname.split('.')[0]
    setOrganizationDomain(organizationDomain)
  }, [setOrganizationDomain])

  const hostname = window.location.hostname
  const parts = hostname.split('.')
  const isLocalDev = hostname === 'localhost' || hostname === '127.0.0.1'
  if (!isLocalDev && parts.length !== 3) {
    return <Navigate to="/404" replace />
  }

  if (isLoading) return <div>Loading org...</div>

  if (error) return <Navigate to="/404" replace />

  return children ? <>{children}</> : <Outlet />
}
