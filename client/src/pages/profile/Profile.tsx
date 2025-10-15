import {
  Button,
  Divider,
  Group,
  Loader,
  Paper,
  Select,
  Stack,
  TextInput,
  Title,
} from '@mantine/core'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import {
  mockOrganization,
  mockUser,
  OrganizationRoles,
  type Organization,
  type User,
} from '../../utils/types'

// TODO: selectors for semantic
// TODO: API
// TODO: notification

export default function Profile() {
  const { userId } = useParams()
  const [user, setUser] = useState<User | null>(null)
  const [isEditing, setEditing] = useState(false)
  const [isAdmin, setIsAdmin] = useState<boolean>(false)
  const [organization, setOrganization] = useState<Organization | null>()
  const [form, setForm] = useState<User | null>(null)
  const currentUserId = '123'

  useEffect(() => {
    // ;(async () => {
    //   const res = await fetch(`${import.meta.env.BASE_URL}/api/users/${userId}`)
    //   const data = await res.json()
    // TODO: API /profile => gets user + organization
    setOrganization(mockOrganization)
    setUser(mockUser)
    setForm(mockUser)
    setIsAdmin(
      mockUser.organizationRole === OrganizationRoles.OWNER ||
        mockUser.organizationRole === OrganizationRoles.ADMIN,
    )
    // })()
  }, [userId])

  if (!user || !form || !organization) return <Loader />

  const isOwnProfile = userId === currentUserId

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    path: string[],
  ) => {
    setForm((prev) => {
      if (!prev) return prev
      const updated: User = structuredClone(prev)
      let target: unknown = updated

      for (let i = 0; i < path.length - 1; i++) {
        const key = path[i] as keyof typeof target
        if (typeof target === 'object' && target !== null)
          target = (target as Record<string, unknown>)[key]
      }

      if (typeof target === 'object' && target !== null) {
        const lastKey = path[path.length - 1]
        ;(target as Record<string, unknown>)[lastKey] = e.target.value
      }

      return updated
    })
  }
  const handleSemanticChange = <K extends keyof User['semantic']>(
    key: K,
    value: string,
  ) => {
    if (!value) return
    setUser((prev) => {
      if (prev == null) return prev
      return { ...prev, semantic: { ...prev.semantic, [key]: value } }
    })
  }

  const handleSave = async () => {
    setEditing(false)
    await fetch(`${import.meta.env.BASE_URL}/api/users/${userId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(form),
    })
    setUser(form)
  }

  return (
    <Paper p="xl" className="max-w-2xl mx-auto mt-10 w-screen">
      <Stack>
        <Group justify="space-between" mt="md">
          <Title order={2}>{form.name}</Title>

          {(isAdmin || isOwnProfile) && (
            <Group justify="flex-end">
              {isEditing ? (
                <>
                  <Button variant="light" onClick={() => setEditing(false)}>
                    Cancel
                  </Button>
                  <Button onClick={handleSave}>Save</Button>
                </>
              ) : (
                <Button bg="green" onClick={() => setEditing(true)}>
                  Edit
                </Button>
              )}
            </Group>
          )}
        </Group>
        <Divider label="General" />
        <TextInput
          label="Email"
          value={form.email}
          readOnly={!isEditing}
          onChange={(e) => handleChange(e, ['email'])}
        />
        <TextInput
          label="Name"
          value={form.name}
          readOnly={!isEditing}
          onChange={(e) => handleChange(e, ['name'])}
        />
        <Divider label="Semantic" />
        <TextInput
          label="Description"
          value={form.semantic.description}
          readOnly={!isEditing}
          onChange={(e) => handleChange(e, ['semantic', 'description'])}
        />

        <Select
          label="Team"
          data={organization.semantic.teams.map((t) => ({
            value: t.id,
            label: t.name,
          }))}
          clearable={false}
          readOnly={!isEditing || !isAdmin}
          value={user.semantic.team}
          onChange={(v) => handleSemanticChange('team', v ?? '')}
        />

        <Select
          label="Level"
          data={organization.semantic.levels.map((l) => ({
            value: l.id,
            label: l.name,
          }))}
          clearable={false}
          readOnly={!isEditing || !isAdmin}
          value={user.semantic.level}
          onChange={(v) => handleSemanticChange('level', v ?? '')}
        />
        <Select
          label="Location"
          data={organization.semantic.locations.map((l) => ({
            value: l,
            label: l,
          }))}
          clearable={false}
          value={user.semantic.location}
          onChange={(v) => handleSemanticChange('location', v ?? '')}
        />
        <Divider label="Contact" />
        <TextInput
          label="Slack"
          value={form.contact.slack}
          readOnly={!isEditing}
          onChange={(e) => handleChange(e, ['contact', 'slack'])}
        />
        <TextInput
          label="Telephone"
          value={form.contact.telephone}
          readOnly={!isEditing}
          onChange={(e) => handleChange(e, ['contact', 'telephone'])}
        />
        <TextInput
          label="Telegram"
          value={form.contact.telegram}
          readOnly={!isEditing}
          onChange={(e) => handleChange(e, ['contact', 'telegram'])}
        />
        <TextInput
          label="WhatsApp"
          value={form.contact.whatsapp}
          readOnly={!isEditing}
          onChange={(e) => handleChange(e, ['contact', 'whatsapp'])}
        />
      </Stack>
    </Paper>
  )
}
