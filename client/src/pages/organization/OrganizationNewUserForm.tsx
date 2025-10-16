import {
  OrganizationRoles,
  UserProviders,
  UserStatuses,
  type Organization,
  type OrganizationRole,
  type User,
} from '@/common/types'
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
import { useState } from 'react'
import { v4 as uuid } from 'uuid'

const defaultUser = (organization: Organization | null): User => {
  if (!organization) throw new Error('No organization')
  return {
    _id: uuid(),
    email: '',
    password: '',
    name: '',
    providers: [UserProviders.Email],
    verified: false,
    status: UserStatuses.ACTIVE,
    semantic: { description: '', team: '', level: '', location: '' },
    contact: {
      slack: '',
      telephone: '',
      email: '',
      telegram: '',
      whatsapp: '',
    },
    avatarId: '',
    organizationId: organization.id,
    organizationRole: OrganizationRoles.USER,
  }
}

type OrganizationNewUserFormProps = {
  onSave: (newUser: User) => void
}
function OrganizationNewUserForm({ onSave }: OrganizationNewUserFormProps) {
  const organization = useOrganizationStore((s) => s.organization)

  const [newUser, setNewUser] = useState<User>(defaultUser(organization))

  const handleSemanticChange = <K extends keyof User['semantic']>(
    key: K,
    value: string,
  ) => {
    setNewUser((prev) => ({
      ...prev,
      semantic: { ...prev.semantic, [key]: value },
    }))
  }

  return (
    <Card withBorder radius="md" shadow="xs" p="md">
      <Stack gap="sm">
        <Group grow>
          <TextInput
            label="Full Name"
            value={newUser.name}
            onChange={(e) =>
              setNewUser((prev) => ({
                ...prev,
                name: e.currentTarget.value,
              }))
            }
          />
          <TextInput
            label="Email"
            value={newUser.email}
            onChange={(e) =>
              setNewUser((prev) => ({
                ...prev,
                email: e.currentTarget.value,
              }))
            }
          />
        </Group>

        <Group grow>
          <Select
            label="Team"
            data={organization?.semantic.teams.map((t) => ({
              value: t.name,
              label: t.name,
            }))}
            value={newUser.semantic.team}
            onChange={(v) => handleSemanticChange('team', v ?? '')}
          />
          <Select
            label="Level"
            data={organization?.semantic.levels.map((l) => ({
              value: l.name,
              label: l.name,
            }))}
            value={newUser.semantic.level}
            onChange={(v) => handleSemanticChange('level', v ?? '')}
          />
        </Group>

        <Textarea
          label="Description"
          value={newUser.semantic.description}
          onChange={(e) =>
            handleSemanticChange('description', e.currentTarget.value)
          }
          minRows={2}
        />

        <Group grow>
          <Select
            label="Role"
            data={Object.values(OrganizationRoles).map((r) => ({
              value: r,
              label: r,
            }))}
            value={newUser.organizationRole}
            onChange={(v) =>
              v &&
              setNewUser((prev) => ({
                ...prev,
                organizationRole: v as OrganizationRole,
              }))
            }
          />
        </Group>

        <Group className="justify-end pt-2">
          <Button
            radius="md"
            onClick={() => {
              onSave(newUser)
              setNewUser(defaultUser(organization))
            }}
          >
            Add User
          </Button>
        </Group>
      </Stack>
    </Card>
  )
}

export default OrganizationNewUserForm
