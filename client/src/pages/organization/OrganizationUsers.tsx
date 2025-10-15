import {
  mockUser,
  type Organization,
  type User,
  UserStatuses,
} from '@/utils/types'
import {
  Badge,
  Button,
  Collapse,
  Container,
  Divider,
  Group,
  Pagination,
  Paper,
  Stack,
  Table,
  TextInput,
  Title,
} from '@mantine/core'
import { useDisclosure, useListState } from '@mantine/hooks'
import { useMemo, useState } from 'react'
import { v4 as uuid } from 'uuid'
import OrganizationNewUserForm from './OrganizationNewUserForm'

type InvitationFormProps = {
  organization: Organization
}

const PAGE_SIZE = 5

export function OrganizationUsers({ organization }: InvitationFormProps) {
  const [users, usersHandlers] = useListState<User>([mockUser])
  const [formOpened, { toggle: toggleForm }] = useDisclosure(false)
  const [search, setSearch] = useState('')
  const [activePage, setActivePage] = useState(1)

  const filteredUsers = useMemo(() => {
    return users.filter(
      (u) =>
        u.name.toLowerCase().includes(search.toLowerCase()) ||
        u.email.toLowerCase().includes(search.toLowerCase()),
    )
  }, [users, search])

  const pageCount = Math.ceil(filteredUsers.length / PAGE_SIZE)
  const paginatedUsers = useMemo(() => {
    const start = (activePage - 1) * PAGE_SIZE
    return filteredUsers.slice(start, start + PAGE_SIZE)
  }, [filteredUsers, activePage])

  const handleAddUser = (newUser: User) => {
    if (!newUser.email || !newUser.name) return
    usersHandlers.append({ ...newUser, _id: uuid() })
  }
  return (
    <Container className="py-12">
      <Paper className="p-8 max-w-5xl mx-auto space-y-8 backdrop-blur-sm">
        <Group className="justify-between items-center">
          <Title order={2}>Organization Users</Title>
        </Group>

        <Divider my="lg" />

        <Stack gap="md">
          <Group
            className="justify-between items-center cursor-pointer select-none"
            onClick={toggleForm}
          >
            <Title order={4}>Invite User</Title>
            <Button variant="light" size="xs">
              {formOpened ? 'Hide' : 'Show'}
            </Button>
          </Group>

          <Collapse in={formOpened}>
            <OrganizationNewUserForm
              organization={organization}
              onSave={handleAddUser}
            />
          </Collapse>
        </Stack>

        <Divider my="lg" label="Current Users" labelPosition="center" />

        <Stack gap="md">
          <Group className="justify-between items-center">
            <TextInput
              placeholder="Search by name or email"
              value={search}
              onChange={(e) => setSearch(e.currentTarget.value)}
              style={{ flex: 1 }}
            />
          </Group>

          <Table.ScrollContainer minWidth={700}>
            <Table highlightOnHover verticalSpacing="sm">
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Name</Table.Th>
                  <Table.Th>Email</Table.Th>
                  <Table.Th>Team</Table.Th>
                  <Table.Th>Level</Table.Th>
                  <Table.Th>Location</Table.Th>
                  <Table.Th>Role</Table.Th>
                  <Table.Th>Status</Table.Th>
                  <Table.Th>Actions</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {paginatedUsers.length === 0 ? (
                  <Table.Tr>
                    <Table.Td
                      colSpan={9}
                      className="text-center py-4 text-gray-500"
                    >
                      No users found
                    </Table.Td>
                  </Table.Tr>
                ) : (
                  paginatedUsers.map((user, idx) => {
                    const statusColor =
                      user.status === UserStatuses.ACTIVE
                        ? 'green'
                        : user.status === UserStatuses.DELETED
                          ? 'red'
                          : 'yellow'

                    return (
                      <Table.Tr key={user._id}>
                        <Table.Td>{user.name}</Table.Td>
                        <Table.Td>{user.email}</Table.Td>
                        <Table.Td>
                          {organization.semantic.teams.find(
                            (t) => t.id === user.semantic.team,
                          )?.name || '—'}
                        </Table.Td>
                        <Table.Td>
                          {organization.semantic.levels.find(
                            (l) => l.id === user.semantic.level,
                          )?.name || '—'}
                        </Table.Td>
                        <Table.Td>{user.semantic.location}</Table.Td>
                        <Table.Td>{user.organizationRole}</Table.Td>
                        <Table.Td>
                          <Badge color={statusColor}>{user.status}</Badge>
                        </Table.Td>
                        <Table.Td>
                          {user.status !== UserStatuses.DELETED ? (
                            <Group>
                              <Button
                                size="xs"
                                color="blue"
                                variant="light"
                                component="a"
                                href={`/profile/${user._id}`}
                              >
                                Edit
                              </Button>
                              <Button
                                size="xs"
                                color="red"
                                variant="light"
                                onClick={() =>
                                  usersHandlers.setItem(idx, {
                                    ...user,
                                    status: UserStatuses.DELETED,
                                  })
                                }
                              >
                                Delete
                              </Button>
                            </Group>
                          ) : (
                            <Button
                              size="xs"
                              color="green"
                              variant="light"
                              onClick={() =>
                                usersHandlers.setItem(idx, {
                                  ...user,
                                  status: UserStatuses.ACTIVE,
                                })
                              }
                            >
                              Restore
                            </Button>
                          )}
                        </Table.Td>
                      </Table.Tr>
                    )
                  })
                )}
              </Table.Tbody>
            </Table>
          </Table.ScrollContainer>

          {pageCount > 1 && (
            <Group justify="flex-end" mt="md">
              <Pagination
                value={activePage}
                onChange={setActivePage}
                total={pageCount}
                size="sm"
              />
            </Group>
          )}
        </Stack>
      </Paper>
    </Container>
  )
}

export default OrganizationUsers
