import { fetchOrganizationUsers } from '@/api/organization'
import { inviteUser, updateUserStatus, type InviteUserData } from '@/api/user'
import { UserStatuses, type UserStatus } from '@/common/types'
import { useAuthStore } from '@/stores/authStore'
import { useOrganizationStore } from '@/stores/organizationStore'
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
import { useDebouncedValue, useDisclosure } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import OrganizationInviteUserForm from './OrganizationInviteUserForm'

const PAGE_SIZE = 5

export function OrganizationUsers() {
  const organization = useOrganizationStore((s) => s.organization)
  const isAdmin = useAuthStore((s) => s.isAdmin)
  const queryClient = useQueryClient()
  const [formOpened, { toggle: toggleForm }] = useDisclosure(false)
  const [search, setSearch] = useState('')
  const [debouncedSearch] = useDebouncedValue(search, 500)

  const [page, setPage] = useState(1)
  const { data, isError, isLoading } = useQuery({
    queryKey: ['organizationUsers', page, debouncedSearch],
    queryFn: () => fetchOrganizationUsers(page, PAGE_SIZE, debouncedSearch),
  })

  const pageCount = Math.ceil((data?.totalCount ?? 0) / PAGE_SIZE)
  const users = data?.users ?? []

  const inviteUserMutation = useMutation({
    mutationFn: inviteUser,
    onError: (error) => {
      console.error('Error inviting user:', error)
      notifications.show({
        title: 'Error',
        message: 'Failed to invite user. Please try again.',
        color: 'red',
      })
    },
    onSuccess: () => {
      notifications.show({
        title: 'Success',
        message: 'User invited successfully',
        color: 'green',
      })
      queryClient.invalidateQueries({
        queryKey: ['organizationUsers', debouncedSearch],
      })
    },
  })

  const updateUserStatusMutation = useMutation({
    mutationFn: ({ userId, status }: { userId: string; status: UserStatus }) =>
      updateUserStatus(userId, status),
    onError: (error) => {
      console.error('Error updating user status:', error)
      notifications.show({
        title: 'Error',
        message: 'Failed to update user status. Please try again.',
        color: 'red',
      })
    },
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: ['organizationUsers', debouncedSearch],
      }),
  })

  const handleAddUser = (newUser: InviteUserData) => {
    if (!newUser.email || !newUser.name) return
    inviteUserMutation.mutate(newUser, {
      onSuccess: () =>
        queryClient.invalidateQueries({
          queryKey: ['organizationUsers', search],
        }),
    })
  }

  const handleChangeStatus = (userId: string, status: UserStatus) => {
    updateUserStatusMutation.mutate({ userId, status })
  }

  return (
    <Container className="py-12">
      <Paper className="p-8 max-w-5xl mx-auto space-y-8 backdrop-blur-sm">
        <Group className="justify-between items-center">
          <Title order={2}>Organization Users</Title>
        </Group>

        <Divider my="lg" />

        {isAdmin && (
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
              <OrganizationInviteUserForm onSave={handleAddUser} />
            </Collapse>
          </Stack>
        )}

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
                {isLoading ? (
                  <Table.Tr>
                    <Table.Td
                      colSpan={9}
                      className="text-center py-6 text-gray-500"
                    >
                      Loading users...
                    </Table.Td>
                  </Table.Tr>
                ) : isError ? (
                  <Table.Tr>
                    <Table.Td
                      colSpan={9}
                      className="text-center py-6 text-red-500"
                    >
                      Failed to load users
                    </Table.Td>
                  </Table.Tr>
                ) : users.length === 0 ? (
                  <Table.Tr>
                    <Table.Td
                      colSpan={9}
                      className="text-center py-4 text-gray-500"
                    >
                      No users found
                    </Table.Td>
                  </Table.Tr>
                ) : (
                  users.map((user) => {
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
                          {organization?.semantic.teams.find(
                            (t) => t.id === user.semantic.team,
                          )?.name || '—'}
                        </Table.Td>
                        <Table.Td>
                          {organization?.semantic.levels.find(
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
                                  handleChangeStatus(
                                    user._id,
                                    UserStatuses.DELETED,
                                  )
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
                                handleChangeStatus(
                                  user._id,
                                  UserStatuses.ACTIVE,
                                )
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
                value={page}
                onChange={setPage}
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
