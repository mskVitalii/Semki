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
        ?.name ?? user.semantic.level,
    [organization?.semantic.levels, user.semantic.level],
  )
  const location = useMemo(
    () =>
      organization?.semantic.locations.find(
        (l) => l.id === user.semantic.location,
      )?.name ?? user.semantic.location,
    [organization?.semantic.locations, user.semantic.location],
  )
  const team = useMemo(
    () =>
      organization?.semantic.teams.find((t) => t.id === user.semantic.team)
        ?.name ?? user.semantic.team,
    [organization?.semantic.teams, user.semantic.team],
  )

  return (
    <Group gap="xs" mt="xs">
      {level && <Badge color="pink">{level}</Badge>}
      {location && (
        <Badge leftSection={<IconMapPin size={12} />} color="orange">
          {location}
        </Badge>
      )}
      {team && <Badge color="purple">{team}</Badge>}
    </Group>
  )
}

export default React.memo(UserBadges)
