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
  console.group('BootstrapRoute')
  const orgQuery = useFetchOrganization()
  const userQuery = useFetchUser()
  console.groupEnd()
  if (orgQuery.isLoading || userQuery.isLoading)
    return (
      <MainLayout>
        <Loader color="green" />
      </MainLayout>
    )

  if (orgQuery.isError || userQuery.isError)
    console.error('error', userQuery.error)

  // if (orgQuery.isError || userQuery.isError)
  //   return <Navigate to="/login" replace />

  return children ? <>{children}</> : <Outlet />
}
