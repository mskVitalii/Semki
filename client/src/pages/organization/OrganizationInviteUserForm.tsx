import type { InviteUserData } from '@/api/user'
import { OrganizationRoles, type Organization } from '@/common/types'
import { useOrganizationStore } from '@/stores/organizationStore'
import {
  Button,
  Card,
  Group,
  Select,
  Stack,
  Textarea,
  TextInput,
} from '@mantine/core'
import { useForm } from '@mantine/form'

const createDefaultUser = (
  organization: Organization | null,
): InviteUserData => {
  if (!organization) throw new Error('No organization')
  return {
    email: '',
    name: '',
    semantic: { team: '', level: '', location: '' },
    organizationRole: OrganizationRoles.USER,
  }
}

type OrganizationInviteUserFormProps = {
  onSave: (newUser: InviteUserData) => void
}

function OrganizationInviteUserForm({
  onSave,
}: OrganizationInviteUserFormProps) {
  const organization = useOrganizationStore((s) => s.organization)

  const form = useForm<InviteUserData>({
    initialValues: createDefaultUser(organization),
    validate: {
      name: (value) => (!value ? 'Name is required' : null),
      email: (value) => {
        if (!value) return 'Email is required'
        if (!/^\S+@\S+\.\S+$/.test(value)) return 'Invalid email format'
        return null
      },
      semantic: {
        team: (value) => (!value ? 'Team is required' : null),
        level: (value) => (!value ? 'Level is required' : null),
      },
    },
  })

  const handleSubmit = (values: InviteUserData) => {
    onSave(values)
    form.reset()
  }

  return (
    <Card withBorder radius="md" shadow="xs" p="md">
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="sm">
          <Group grow>
            <TextInput
              label="Full Name"
              placeholder="Enter full name"
              withAsterisk
              {...form.getInputProps('name')}
            />
            <TextInput
              label="Email"
              placeholder="user@example.com"
              withAsterisk
              {...form.getInputProps('email')}
            />
          </Group>

          <Group grow>
            <Select
              label="Team"
              placeholder="Select team"
              withAsterisk
              data={
                organization?.semantic.teams.map((t) => ({
                  value: t.name,
                  label: t.name,
                })) ?? []
              }
              {...form.getInputProps('semantic.team')}
            />
            <Select
              label="Level"
              placeholder="Select level"
              withAsterisk
              data={
                organization?.semantic.levels.map((l) => ({
                  value: l.name,
                  label: l.name,
                })) ?? []
              }
              {...form.getInputProps('semantic.level')}
            />
          </Group>

          <Textarea
            label="Description"
            placeholder="Enter user description"
            minRows={2}
            {...form.getInputProps('semantic.description')}
          />

          <Group grow>
            <Select
              label="Role"
              placeholder="Select role"
              data={Object.values(OrganizationRoles).map((r) => ({
                value: r,
                label: r,
              }))}
              {...form.getInputProps('organizationRole')}
            />
          </Group>

          <Group justify="flex-end" pt="md">
            <Button variant="subtle" onClick={() => form.reset()} type="button">
              Reset
            </Button>
            <Button radius="md" type="submit">
              Add User
            </Button>
          </Group>
        </Stack>
      </form>
    </Card>
  )
}

export default OrganizationInviteUserForm
