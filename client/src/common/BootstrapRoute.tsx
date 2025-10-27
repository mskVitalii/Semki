import { Loader } from '@mantine/core'
import { Outlet } from 'react-router-dom'
import { MainLayout } from './SidebarLayout'
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
    return (
      <MainLayout>
        <Loader color="green" />
      </MainLayout>
    )

  if (orgQuery.isError || userQuery.isError)
    console.error('error', userQuery.error)
  //   return <Navigate to="/login" replace />

  return children ? <>{children}</> : <Outlet />
}
