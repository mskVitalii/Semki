/* eslint-disable @typescript-eslint/no-explicit-any */
import { api } from '@/api/client'
import { MainLayout } from '@/common/SidebarLayout'
import { type User } from '@/common/types'
import UserBadges from '@/common/UserBadges'
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
import ProfileReadOnly from './ProfileReadOnly'

function diffUser(a: User, b: User): Partial<User> {
  function compare(valA: any, valB: any): any {
    if (typeof valA !== 'object' || valA === null) {
      return valA !== valB ? valB : undefined
    }

    const result: any = Array.isArray(valA) ? [] : {}
    let hasDiff = false

    for (const key of Object.keys(valA)) {
      if (valB && key in valB) {
        const diff = compare(valA[key], valB[key])
        if (diff !== undefined) {
          result[key] = diff
          hasDiff = true
        }
      } else {
        result[key] = valB?.[key]
        hasDiff = true
      }
    }

    return hasDiff ? result : undefined
  }

  return compare(a, b) ?? {}
}

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
    mutationFn: (payload: Partial<User>) => {
      console.log('mutation', payload)
      return api.patch(`/api/v1/user/${userId}`, payload)
    },
    onMutate: () => setStatus('saving'),
    onSuccess: () => {
      setStatus('saved')
      setTimeout(() => setStatus('idle'), 1500)
      queryClient.invalidateQueries({ queryKey: ['user', userId] })
      if (userId === currUserId)
        queryClient.invalidateQueries({ queryKey: ['user'] })
      console.log('onSuccess', user)
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
    mode: 'controlled',
    initialValues: user || ({} as User),
  })

  useEffect(() => {
    if (!user || !user._id || !userId || isLoading) return
    console.log('useEffect reset')
    form.setValues(user)
    form.setInitialValues(user)
    form.reset()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user])

  const [debounced] = useDebouncedValue(form.values, 2000)

  useEffect(() => {
    if (!user || !user._id || !userId || !debounced) return
    if (Object.keys(debounced).length === 0) return
    const changes = diffUser(user, debounced)
    if (Object.keys(changes).length === 0) return
    console.log('changes', changes, debounced, form.values)
    mutation.mutate(changes)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debounced, userId])

  if (
    !user ||
    !organization ||
    isLoading ||
    Object.keys(form.values).length === 0
  )
    return (
      <MainLayout>
        <div className="flex items-center h-full">
          <Loader color="green" />
        </div>
      </MainLayout>
    )

  if (!isAdmin && userId !== currUserId) return <ProfileReadOnly user={user} />

  return (
    <MainLayout>
      <Card
        shadow="lg"
        padding="xl"
        radius="md"
        m="xl"
        className="min-h-screen w-full p-4 overflow-auto!"
      >
        <Stack>
          <Group justify="space-between" mt="md">
            <Title order={2}>{form.values.name}</Title>
          </Group>
          <UserBadges user={user} />

          <Divider label="General" m="lg" />
          <TextInput
            label="Email"
            readOnly={!isAdmin}
            styles={{ label: { marginBottom: '0.75rem' } }}
            key={form.key('email')}
            {...form.getInputProps('email')}
          />
          <TextInput
            label="Name"
            styles={{ label: { marginBottom: '0.75rem' } }}
            key={form.key('name')}
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
            key={form.key('semantic.description')}
            {...form.getInputProps('semantic.description')}
          />

          <Select
            label="Location"
            allowDeselect={false}
            styles={{ label: { marginBottom: '0.75rem' } }}
            data={organization.semantic.locations.map((l) => ({
              value: l.id!,
              label: l.name,
            }))}
            key={form.key('semantic.location')}
            {...form.getInputProps('semantic.location')}
          />

          <Group grow>
            <Select
              label="Team"
              allowDeselect={false}
              readOnly={!isAdmin}
              styles={{ label: { marginBottom: '0.75rem' } }}
              data={organization.semantic.teams.map((t) => ({
                value: t.id!,
                label: t.name,
              }))}
              key={form.key('semantic.team')}
              {...form.getInputProps('semantic.team')}
            />
            <Select
              label="Level"
              readOnly={!isAdmin}
              allowDeselect={false}
              styles={{ label: { marginBottom: '0.75rem' } }}
              data={organization.semantic.levels.map((l) => ({
                value: l.id!,
                label: l.name,
              }))}
              key={form.key('semantic.level')}
              {...form.getInputProps('semantic.level')}
            />
          </Group>
          <Divider label="Contact" m="lg" />
          <Group grow>
            <TextInput
              label="Slack"
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('semantic.slack')}
              {...form.getInputProps('contact.slack')}
            />
            <TextInput
              label="Telephone"
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('semantic.telephone')}
              {...form.getInputProps('contact.telephone')}
            />
          </Group>
          <Group grow>
            <TextInput
              label="Telegram"
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('semantic.telegram')}
              {...form.getInputProps('contact.telegram')}
            />
            <TextInput
              label="WhatsApp"
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('semantic.whatsapp')}
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
