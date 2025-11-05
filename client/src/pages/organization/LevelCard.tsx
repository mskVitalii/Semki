import { api } from '@/api/client'
import { type Level } from '@/common/types'
import { ActionIcon, Card, Stack, TextInput, Textarea } from '@mantine/core'
import { useDebouncedCallback } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'
import { IconTrash } from '@tabler/icons-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'

export function LevelCard({
  level,
  onChange,
  onDelete,
  disabled,
}: {
  level: Level
  onDelete: () => void
  onChange: (lvl: Level) => void
  disabled?: boolean
}) {
  const queryClient = useQueryClient()

  const [status, setStatus] = useState<'idle' | 'save' | 'saved' | 'delete'>(
    'idle',
  )

  const mutation = useMutation({
    mutationFn: ({
      lvl,
      method,
    }: {
      lvl: Level
      method: 'save' | 'delete'
    }) => {
      if (method === 'delete') {
        return api.delete(`/api/v1/organization/levels/${lvl.id}`)
      }

      return lvl.id
        ? api.put(`/api/v1/organization/levels/${lvl.id}`, lvl)
        : api.post(`/api/v1/organization/levels`, lvl)
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

  const debouncedSave = useDebouncedCallback((lvl: Level) => {
    if (disabled) return
    if (level.name.trim() === '') return
    if (level.description.trim() === '') return
    mutation.mutate({ lvl, method: 'save' })
  }, 1000)

  const handleDelete = (lvl: Level) => {
    if (!lvl.id) {
      onDelete()
      return
    }
    mutation.mutate({ lvl, method: 'delete' })
  }

  const handleChange = (key: keyof Level, value: string) => {
    const updated = { ...level, [key]: value }
    onChange(updated)
    debouncedSave(updated)
  }

  return (
    <Card withBorder radius="md" shadow="xs" p="md">
      {!disabled && (
        <ActionIcon
          color="red"
          size="sm"
          onClick={() => handleDelete(level)}
          className="absolute! top-2 right-2"
        >
          <IconTrash size={16} />
        </ActionIcon>
      )}
      <Stack gap="xs">
        <TextInput
          label="Level Name"
          value={level.name}
          styles={{ label: { marginBottom: '0.75rem' } }}
          disabled={disabled}
          onChange={(e) => handleChange('name', e.currentTarget.value)}
        />
        <Textarea
          label="Level Description"
          minRows={2}
          value={level.description}
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
