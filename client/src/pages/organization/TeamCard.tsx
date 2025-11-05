import { api } from '@/api/client'
import { type Team } from '@/common/types'
import { ActionIcon, Card, Stack, TextInput, Textarea } from '@mantine/core'
import { useDebouncedCallback } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'
import { IconTrash } from '@tabler/icons-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'

export function TeamCard({
  team,
  onChange,
  onDelete,
  disabled,
}: {
  team: Team
  onChange: (team: Team) => void
  onDelete: () => void
  disabled?: boolean
}) {
  const [status, setStatus] = useState<'idle' | 'save' | 'saved' | 'delete'>(
    'idle',
  )
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: ({
      team,
      method,
    }: {
      team: Team
      method: 'save' | 'delete'
    }) => {
      if (method === 'delete') {
        return api.delete(`/api/v1/organization/teams/${team.id}`)
      }

      return team.id
        ? api.put(`/api/v1/organization/teams/${team.id}`, team)
        : api.post(`/api/v1/organization/teams`, team)
    },
    onMutate: ({ method }) => setStatus(method),
    onSuccess: (_, { method }) => {
      if (method !== 'delete') {
        setStatus('saved')
        setTimeout(() => setStatus('idle'), 1500)
      } else {
        setStatus('idle')
      }
      queryClient.invalidateQueries({
        queryKey: ['organization'],
      })
    },
    onError: () => {
      notifications.show({
        title: 'Error',
        message: 'Failed to save',
        color: 'red',
      })
      setStatus('idle')
    },
  })

  const debouncedSave = useDebouncedCallback((team: Team) => {
    if (team.name.trim() === '') return
    if (team.description.trim() === '') return
    mutation.mutate({ team, method: 'save' })
  }, 1000)

  const handleDelete = (team: Team) => {
    if (!team.id) {
      onDelete()
      return
    }
    mutation.mutate({ team, method: 'delete' })
  }

  const handleChange = (key: keyof Team, value: string) => {
    const updated = { ...team, [key]: value }
    onChange(updated)
    debouncedSave(updated)
  }

  return (
    <Card withBorder radius="md" shadow="xs" p="md" className="relative">
      {!disabled && (
        <ActionIcon
          color="red"
          size="sm"
          onClick={() => handleDelete(team)}
          className="absolute! top-2 right-2"
        >
          <IconTrash size={16} />
        </ActionIcon>
      )}
      <Stack gap="xs">
        <TextInput
          label="Team Name"
          value={team.name}
          disabled={disabled}
          styles={{ label: { marginBottom: '0.75rem' } }}
          onChange={(e) => handleChange('name', e.currentTarget.value)}
        />
        <Textarea
          label="Team Description"
          minRows={2}
          value={team.description}
          disabled={disabled}
          styles={{ label: { marginBottom: '0.75rem' } }}
          onChange={(e) => handleChange('description', e.currentTarget.value)}
        />
        {status !== 'idle' && status !== 'delete' && (
          <div className="text-sm text-gray-500">
            {status === 'save' ? 'Saving…' : 'Saved'}
          </div>
        )}
      </Stack>
    </Card>
  )
}
