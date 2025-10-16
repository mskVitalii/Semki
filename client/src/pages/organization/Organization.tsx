import { MainLayout } from '@/common/SidebarLayout'
import { Paper } from '@mantine/core'
import OrganizationSettings from './OrganizationSettings'
import OrganizationUsers from './OrganizationUsers'

function Organization() {
  return (
    <MainLayout>
      <Paper p="xl" className="mt-10 w-full space-y-6!">
        <OrganizationSettings />
        <OrganizationUsers />
      </Paper>
    </MainLayout>
  )
}

export default Organization
