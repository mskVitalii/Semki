import { api } from '@/api/client'
import type { Location } from '@/common/types'
import { useOrganizationStore } from '@/stores/organizationStore'
import { Button, Collapse, Group, Stack, TagsInput, Title } from '@mantine/core'
import { useDisclosure, useListState } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'

export function LocationSection({ disabled }: { disabled?: boolean }) {
  const [opened, { toggle }] = useDisclosure(false)
  const [status, setStatus] = useState<'idle' | 'saving' | 'saved'>('idle')
  const queryClient = useQueryClient()

  const organization = useOrganizationStore((s) => s.organization)
  const [locations, locationsHandlers] = useListState<Location>(
    organization?.semantic.locations ?? [],
  )

  const mutation = useMutation({
    mutationFn: ({
      location,
      method,
    }: {
      location: Location
      method: 'save' | 'delete'
    }) => {
      if (method === 'delete') {
        if (!location.id)
          throw new Error('Location ID is required for deletion')
        return api.delete(`/api/v1/organization/locations/${location.id}`)
      }
      return location.id
        ? api.put(`/api/v1/organization/locations/${location.id}`, location)
        : api.post(`/api/v1/organization/locations`, location)
    },
    onMutate: () => setStatus('saving'),
    onSuccess: () => {
      setStatus('saved')
      setTimeout(() => setStatus('idle'), 1500)
      queryClient.invalidateQueries({ queryKey: ['organization'] })
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

  const handleChange = (values: string[]) => {
    const newLocations: Location[] = values.map((name) => {
      const existing = locations.find((l) => l.name === name)
      if (existing) return existing
      const newLoc = { id: undefined, name }
      mutation.mutate({ location: newLoc, method: 'save' })
      return newLoc
    })

    locations.forEach((loc) => {
      if (!values.includes(loc.name) && loc.id) {
        mutation.mutate({ location: loc, method: 'delete' })
      }
    })

    locationsHandlers.setState(newLocations)
  }

  return (
    <Stack gap="md">
      <Group
        className="justify-between items-center cursor-pointer select-none"
        onClick={toggle}
      >
        <Title order={4}>Locations</Title>
        <Button variant="light" size="xs">
          {opened ? 'Hide' : 'Show'}
        </Button>
      </Group>
      <Collapse in={opened}>
        <TagsInput
          value={locations.map((l) => l.name)}
          disabled={disabled}
          onChange={handleChange}
          placeholder="Add location..."
          radius="md"
          size="md"
          className="pt-2"
        />
        {status !== 'idle' && (
          <div className="text-sm text-gray-500">
            {status === 'saving' ? 'Saving…' : 'Saved'}
          </div>
        )}
      </Collapse>
    </Stack>
  )
}
