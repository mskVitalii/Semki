import { MainLayout } from '@/common/SidebarLayout'
import { type User } from '@/common/types'
import { useOrganizationStore } from '@/stores/organizationStore'
import { useUserStore } from '@/stores/userStore'
import {
  Button,
  Card,
  Divider,
  Group,
  Loader,
  Select,
  Stack,
  TextInput,
  Title,
} from '@mantine/core'
import { useState } from 'react'
import { useParams } from 'react-router-dom'

// TODO: API
// TODO: notification

export default function Profile() {
  const { userId } = useParams()
  const [isEditing, setEditing] = useState(false)
  const isAdmin = useUserStore((s) => s.isAdmin)
  const user = useUserStore((s) => s.user)
  const organization = useOrganizationStore((s) => s.organization)
  const [form, setForm] = useState<User | null>(user)

  if (!form || !organization)
    return (
      <MainLayout>
        <Loader color="green" />
      </MainLayout>
    )

  const isOwnProfile = userId === user?._id

  const handleChangeEvent = (
    e: React.ChangeEvent<HTMLInputElement>,
    path: string[],
  ) => {
    handleChange(e.target.value, path)
  }

  const handleChange = (value: string, path: string[]) => {
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
        ;(target as Record<string, unknown>)[lastKey] = value
      }

      return updated
    })
  }

  const handleSave = async () => {
    setEditing(false)
    // TODO: change to tanstack
    if (!userId) return
    await fetch(`${import.meta.env.VITE_API_URL}/api/v1/users/${userId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(form),
    })
  }

  return (
    <MainLayout>
      <Card
        shadow="lg"
        padding="xl"
        radius="md"
        className="min-h-screen w-full p-4 overflow-auto!"
      >
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
            onChange={(e) => handleChangeEvent(e, ['email'])}
          />
          <TextInput
            label="Name"
            value={form.name}
            readOnly={!isEditing}
            onChange={(e) => handleChangeEvent(e, ['name'])}
          />
          <Divider label="Semantic" />
          <TextInput
            label="Description"
            value={form.semantic.description}
            readOnly={!isEditing}
            onChange={(e) => handleChangeEvent(e, ['semantic', 'description'])}
          />

          <Select
            label="Team"
            data={organization.semantic.teams.map((t) => ({
              value: t.id!,
              label: t.name,
            }))}
            clearable={false}
            readOnly={!isEditing || !isAdmin}
            value={form.semantic.team}
            onChange={(v) => handleChange(v ?? '', ['semantic', 'team'])}
          />

          <Select
            label="Level"
            data={organization.semantic.levels.map((l) => ({
              value: l.id!,
              label: l.name,
            }))}
            clearable={false}
            readOnly={!isEditing || !isAdmin}
            value={form.semantic.level}
            onChange={(v) => handleChange(v ?? '', ['semantic', 'level'])}
          />
          <Select
            label="Location"
            data={organization.semantic.locations.map((l) => ({
              value: l.id!,
              label: l.name,
            }))}
            clearable={false}
            value={form.semantic.location}
            onChange={(v) => handleChange(v ?? '', ['semantic', 'location'])}
          />
          <Divider label="Contact" />
          <TextInput
            label="Slack"
            value={form.contact.slack}
            readOnly={!isEditing}
            onChange={(e) => handleChangeEvent(e, ['contact', 'slack'])}
          />
          <TextInput
            label="Telephone"
            value={form.contact.telephone}
            readOnly={!isEditing}
            onChange={(e) => handleChangeEvent(e, ['contact', 'telephone'])}
          />
          <TextInput
            label="Telegram"
            value={form.contact.telegram}
            readOnly={!isEditing}
            onChange={(e) => handleChangeEvent(e, ['contact', 'telegram'])}
          />
          <TextInput
            label="WhatsApp"
            value={form.contact.whatsapp}
            readOnly={!isEditing}
            onChange={(e) => handleChangeEvent(e, ['contact', 'whatsapp'])}
          />
        </Stack>
      </Card>
    </MainLayout>
  )
}
