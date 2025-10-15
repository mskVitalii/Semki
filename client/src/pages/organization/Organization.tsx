import { Paper } from '@mantine/core'
import { mockOrganization } from '../types'
import OrganizationSettings from './OrganizationSettings'
import Users from './OrganizationUsers'

function Organization() {
  // TODO: API organization
  return (
    <Paper p="xl" className="mt-10 w-screen space-y-6!">
      <OrganizationSettings />
      <Users organization={mockOrganization} />
    </Paper>
  )
}

export default Organization
