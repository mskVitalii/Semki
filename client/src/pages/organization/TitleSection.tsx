import { api } from '@/api/client'
import { useOrganizationStore } from '@/stores/organizationStore'
import { TextInput } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'

export function TitleSection({ disabled }: { disabled?: boolean }) {
  const organization = useOrganizationStore((s) => s.organization)
  const queryClient = useQueryClient()

  const [title, setTitle] = useState(organization?.title ?? '')
  const [debouncedTitle] = useDebouncedValue(title, 500)
  const [status, setStatus] = useState<'idle' | 'saving' | 'saved'>('idle')

  const mutation = useMutation({
    mutationFn: (title: string) => api.patch(`/api/v1/organization`, { title }),
    onMutate: () => setStatus('saving'),
    onSuccess: () => {
      setStatus('saved')
      setTimeout(() => setStatus('idle'), 1500)
      queryClient.invalidateQueries({ queryKey: ['organization'] })
    },
    onError: () => {
      notifications.show({
        title: 'Error',
        message: 'Failed to update title',
        color: 'red',
      })
      setStatus('idle')
    },
  })

  useEffect(() => {
    if (!organization) return
    if (debouncedTitle === organization.title) return
    mutation.mutate(debouncedTitle)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedTitle])

  return (
    <>
      <TextInput
        label="Title"
        value={title}
        disabled={disabled}
        onChange={(e) => setTitle(e.currentTarget.value)}
        radius="md"
        size="md"
      />
      {status !== 'idle' && (
        <div className="text-sm text-gray-500">
          {status === 'saving' ? 'Saving…' : 'Saved'}
        </div>
      )}
    </>
  )
}
