import { api } from '@/api/client'
import { MainLayout } from '@/common/SidebarLayout'
import { type User } from '@/common/types'
import { useAuthStore } from '@/stores/authStore'
import { useOrganizationStore } from '@/stores/organizationStore'
import {
  Card,
  Divider,
  Group,
  Loader,
  Select,
  Stack,
  Textarea,
  TextInput,
  Title,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { useDebouncedValue } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'

export default function Profile() {
  const { userId } = useParams<{ userId: string }>()
  const { data: user, isLoading } = useQuery({
    queryKey: ['user', userId],
    queryFn: async () => {
      const { data } = await api.get<User>(`/api/v1/user/${userId}`)
      return data
    },
    enabled: !!userId,
  })

  const isAdmin = useAuthStore((s) => s.isAdmin)
  const currUserId = useAuthStore((s) => s.claims?._id)

  const queryClient = useQueryClient()
  const organization = useOrganizationStore((s) => s.organization)
  const [status, setStatus] = useState<'idle' | 'saving' | 'saved'>('idle')

  const mutation = useMutation({
    mutationFn: (payload: Partial<User>) =>
      api.patch(`/api/v1/user/${userId}`, payload),
    onMutate: () => setStatus('saving'),
    onSuccess: () => {
      setStatus('saved')
      setTimeout(() => setStatus('idle'), 1500)
      queryClient.invalidateQueries({ queryKey: ['user', userId] })
      if (userId === currUserId)
        queryClient.invalidateQueries({ queryKey: ['user'] })
    },
    onError: () => {
      notifications.show({
        title: 'Error',
        message: 'Failed to update profile',
        color: 'red',
      })
      setStatus('idle')
    },
  })

  const form = useForm<User>({
    mode: 'uncontrolled',
    initialValues: user || ({} as User),
  })

  useEffect(() => {
    if (user) {
      form.setValues(user)
      form.setInitialValues(user)
      form.reset()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user])

  const [debounced] = useDebouncedValue(form.values, 1000)
  const [isReady, setIsReady] = useState(false)

  useEffect(() => {
    if (isLoading) return
    if (!isReady) {
      setIsReady(true)
      return
    }
    if (userId) mutation.mutate(debounced)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debounced, userId])

  if (!user || !organization || isLoading)
    return (
      <MainLayout>
        <div className="flex items-center h-full">
          <Loader color="green" />
        </div>
      </MainLayout>
    )

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
            <Title order={2}>{form.values.name}</Title>
          </Group>
          <Divider label="General" m="lg" />
          <TextInput
            label="Email"
            readOnly={!isAdmin}
            styles={{ label: { marginBottom: '0.75rem' } }}
            {...form.getInputProps('email')}
          />
          <TextInput
            label="Name"
            styles={{ label: { marginBottom: '0.75rem' } }}
            {...form.getInputProps('name')}
          />
          <Divider label="Semantic" m="lg" />
          <Textarea
            label="Description"
            placeholder="Enter user description"
            minRows={form.values.semantic?.description?.split('\n').length ?? 4}
            styles={{
              label: { marginBottom: '0.75rem' },
              input: { resize: 'vertical' },
            }}
            {...form.getInputProps('semantic.description')}
          />

          <Select
            label="Location"
            styles={{ label: { marginBottom: '0.75rem' } }}
            data={organization.semantic.locations.map((l) => ({
              value: l.id!,
              label: l.name,
            }))}
            {...form.getInputProps('semantic.location')}
          />

          <Group grow>
            <Select
              label="Team"
              readOnly={!isAdmin}
              styles={{ label: { marginBottom: '0.75rem' } }}
              data={organization.semantic.teams.map((t) => ({
                value: t.id!,
                label: t.name,
              }))}
              {...form.getInputProps('semantic.team')}
            />
            <Select
              label="Level"
              readOnly={!isAdmin}
              styles={{ label: { marginBottom: '0.75rem' } }}
              data={organization.semantic.levels.map((l) => ({
                value: l.id!,
                label: l.name,
              }))}
              {...form.getInputProps('semantic.level')}
            />
          </Group>
          <Divider label="Contact" m="lg" />
          <Group grow>
            <TextInput
              label="Slack"
              styles={{ label: { marginBottom: '0.75rem' } }}
              {...form.getInputProps('contact.slack')}
            />
            <TextInput
              label="Telephone"
              styles={{ label: { marginBottom: '0.75rem' } }}
              {...form.getInputProps('contact.telephone')}
            />
          </Group>
          <Group grow>
            <TextInput
              label="Telegram"
              styles={{ label: { marginBottom: '0.75rem' } }}
              {...form.getInputProps('contact.telegram')}
            />
            <TextInput
              label="WhatsApp"
              styles={{ label: { marginBottom: '0.75rem' } }}
              {...form.getInputProps('contact.whatsapp')}
            />
          </Group>
          {status !== 'idle' && (
            <div className="text-sm text-gray-500">
              {status === 'saving' ? 'Saving…' : 'Saved'}
            </div>
          )}
        </Stack>
      </Card>
    </MainLayout>
  )
}
