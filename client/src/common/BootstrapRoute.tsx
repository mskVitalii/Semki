import { Navigate, Outlet } from 'react-router-dom'
import { useFetchOrganization } from './useFetchOrganization'
import { useFetchUser } from './useFetchUser'

export const BootstrapRoute = ({
  children,
}: {
  children?: React.ReactNode
}) => {
  const orgQuery = useFetchOrganization()
  const userQuery = useFetchUser()

  if (orgQuery.isLoading || userQuery.isLoading)
    return <div>Loading org...</div>

  if (orgQuery.isError || userQuery.isError)
    return <Navigate to="/404" replace />

  return children ? <>{children}</> : <Outlet />
}
