import { type Level } from '@/common/types'
import { useOrganizationStore } from '@/stores/organizationStore'
import { Button, Collapse, Group, Stack, Title } from '@mantine/core'
import { useDisclosure, useListState } from '@mantine/hooks'
import { LevelCard } from './LevelCard'

export function LevelSection({ disabled }: { disabled?: boolean }) {
  const organization = useOrganizationStore((s) => s.organization)

  const [opened, { toggle }] = useDisclosure(false)
  const [levels, handlers] = useListState<Level>(
    organization?.semantic.levels ?? [],
  )

  const handleChange = (idx: number, updated: Level) => {
    handlers.setItem(idx, updated)
  }

  const addLevel = () => {
    handlers.append({ name: '', description: '' })
  }
  const deleteLevel = (idx: number) => {
    handlers.remove(idx)
  }

  return (
    <Stack gap="md">
      <Group
        className="justify-between items-center cursor-pointer select-none"
        onClick={toggle}
      >
        <Title order={4}>Levels</Title>
        <Button variant="light" size="xs">
          {opened ? 'Hide' : 'Show'}
        </Button>
      </Group>
      <Collapse in={opened}>
        <Stack gap="md" className="pt-2">
          {levels.map((level, i) => (
            <LevelCard
              key={i}
              level={level}
              onDelete={() => deleteLevel(i)}
              onChange={(lvl) => handleChange(i, lvl)}
              disabled={disabled}
            />
          ))}
          {!disabled && (
            <Button
              size="sm"
              variant="outline"
              onClick={addLevel}
              disabled={disabled || levels.some((x) => x.name.trim() === '')}
            >
              Add Level
            </Button>
          )}
        </Stack>
      </Collapse>
    </Stack>
  )
}
