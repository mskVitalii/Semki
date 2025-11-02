import { type Team } from '@/common/types'
import { useOrganizationStore } from '@/stores/organizationStore'
import { Button, Collapse, Group, Stack, Title } from '@mantine/core'
import { useDisclosure, useListState } from '@mantine/hooks'
import { TeamCard } from './TeamCard'

export function TeamSection({ disabled }: { disabled?: boolean }) {
  const organization = useOrganizationStore((s) => s.organization)
  const [opened, { toggle }] = useDisclosure(false)
  const [teams, handlers] = useListState<Team>(
    organization?.semantic.teams ?? [],
  )

  const handleChange = (idx: number, updated: Team) => {
    handlers.setItem(idx, updated)
  }
  const addTeam = () => {
    handlers.append({ name: '', description: '' })
  }
  const deleteTeam = (idx: number) => {
    handlers.remove(idx)
  }

  return (
    <Stack gap="md">
      <Group
        className="justify-between items-center cursor-pointer select-none"
        onClick={toggle}
      >
        <Title order={4}>Teams</Title>
        <Button variant="light" size="xs">
          {opened ? 'Hide' : 'Show'}
        </Button>
      </Group>
      <Collapse in={opened}>
        <Stack gap="md" className="pt-2">
          {teams.map((team, i) => (
            <TeamCard
              key={i}
              team={team}
              onDelete={() => deleteTeam(i)}
              onChange={(t) => handleChange(i, t)}
              disabled={disabled}
            />
          ))}
          {!disabled && (
            <Button
              size="sm"
              variant="outline"
              onClick={addTeam}
              disabled={disabled || teams.some((x) => x.name.trim() === '')}
            >
              Add Team
            </Button>
          )}
        </Stack>
      </Collapse>
    </Stack>
  )
}
