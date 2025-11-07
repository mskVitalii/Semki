import { useOrganizationStore } from '@/stores/organizationStore'
import { Badge, Group } from '@mantine/core'
import { IconMapPin } from '@tabler/icons-react'
import React, { useMemo } from 'react'
import type { User } from './types'

export function UserBadges({ user }: { user: User }) {
  const organization = useOrganizationStore((s) => s.organization)
  const level = useMemo(
    () =>
      organization?.semantic.levels.find((l) => l.id === user.semantic.level)
        ?.name,
    [organization?.semantic.levels, user.semantic.level],
  )
  const location = useMemo(
    () =>
      organization?.semantic.locations.find(
        (l) => l.id === user.semantic.location,
      )?.name,
    [organization?.semantic.locations, user.semantic.location],
  )
  const team = useMemo(
    () =>
      organization?.semantic.teams.find((t) => t.id === user.semantic.team)
        ?.name,
    [organization?.semantic.teams, user.semantic.team],
  )

  return (
    <Group gap="xs" mt="xs">
      {level && <Badge color="blue">{level}</Badge>}
      {location && (
        <Badge leftSection={<IconMapPin size={12} />} color="green">
          {location}
        </Badge>
      )}
      {team && <Badge color="violet">{team}</Badge>}
    </Group>
  )
}

export default React.memo(UserBadges)
