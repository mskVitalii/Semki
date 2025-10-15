import {
  mockOrganization,
  mockUser,
  OrganizationRoles,
  type Level,
  type Organization,
  type Team,
} from '@/utils/types'
import {
  Button,
  Card,
  Collapse,
  Container,
  Divider,
  Group,
  Paper,
  Stack,
  TagsInput,
  Textarea,
  TextInput,
  Title,
} from '@mantine/core'
import { useDisclosure, useListState } from '@mantine/hooks'
import { useEffect, useState } from 'react'
import { v4 as uuid } from 'uuid'

export function OrganizationSettings() {
  const [organization, setOrganization] = useState<Organization | null>(null)
  const [title, setTitle] = useState('')
  const [levels, levelsHandlers] = useListState<Level>([])
  const [locations, locationsHandlers] = useListState<string>([])
  const [teams, teamsHandlers] = useListState<Team>([])

  const [levelsOpened, { toggle: toggleLevels }] = useDisclosure(false)
  const [locationsOpened, { toggle: toggleLocations }] = useDisclosure(false)
  const [teamsOpened, { toggle: toggleTeams }] = useDisclosure(false)

  const addLevel = () => {
    levelsHandlers.append({
      id: uuid(),
      name: '',
      description: '',
    })
  }

  const addTeam = () => {
    console.log('Add data', uuid())
    teamsHandlers.append({
      id: uuid(),
      name: '',
      description: '',
    })
  }

  const canEdit =
    mockUser.organizationRole === OrganizationRoles.OWNER ||
    mockUser.organizationRole === OrganizationRoles.ADMIN

  useEffect(() => {
    console.log('useEffect')
    // TODO: API
    setOrganization(mockOrganization)
    setTitle(mockOrganization.title)
    levelsHandlers.setState(mockOrganization.semantic.levels)
    locationsHandlers.setState(mockOrganization.semantic.locations)
    teamsHandlers.setState(mockOrganization.semantic.teams)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleSave = () => {
    // TODO: API
    if (!organization) return
    setOrganization({
      ...organization,
      title,
      semantic: {
        ...organization.semantic,
        levels,
        locations,
        teams,
      },
    })
    // TODO: save notification
  }

  if (!organization) return <div>Loading...</div>

  return (
    <Container className="py-12!">
      <Paper className="p-8! max-w-3xl! mx-auto! space-y-8!">
        <Group className="justify-between items-center">
          <Title order={2}>Organization Settings</Title>
        </Group>

        <Stack gap="md">
          <TextInput
            label="Organization Title"
            value={title}
            onChange={(e) => setTitle(e.currentTarget.value)}
            radius="md"
            size="md"
          />
          <TextInput
            label="Current Plan"
            value={organization.plan}
            disabled
            radius="md"
            size="md"
          />
        </Stack>

        <Divider my="lg" />

        <Stack gap="md">
          <Group
            className="justify-between items-center cursor-pointer select-none"
            onClick={toggleLevels}
          >
            <Title order={4}>Levels</Title>
            <Button variant="light" size="xs">
              {levelsOpened ? 'Hide' : 'Show'}
            </Button>
          </Group>
          <Collapse in={levelsOpened}>
            <Stack gap="md" className="pt-2">
              {levels.map((level, idx) => (
                <Card key={level.id} withBorder radius="md" shadow="xs" p="md">
                  <Stack gap="xs">
                    <TextInput
                      label="Level Name"
                      value={level.name}
                      onChange={(e) =>
                        levelsHandlers.setItem(idx, {
                          ...level,
                          name: e.currentTarget.value,
                        })
                      }
                    />
                    <Textarea
                      label="Level Description"
                      minRows={2}
                      value={level.description}
                      onChange={(e) =>
                        levelsHandlers.setItem(idx, {
                          ...level,
                          description: e.currentTarget.value,
                        })
                      }
                    />
                  </Stack>
                </Card>
              ))}
              {canEdit && (
                <Button size="sm" variant="outline" onClick={addLevel}>
                  Add Level
                </Button>
              )}
            </Stack>
          </Collapse>
        </Stack>

        <Divider my="lg" />

        <Stack gap="md">
          <Group
            className="justify-between items-center cursor-pointer select-none"
            onClick={toggleLocations}
          >
            <Title order={4}>Locations</Title>
            <Button variant="light" size="xs">
              {locationsOpened ? 'Hide' : 'Show'}
            </Button>
          </Group>
          <Collapse in={locationsOpened}>
            <TagsInput
              label="Locations"
              value={locations}
              onChange={locationsHandlers.setState}
              placeholder="Add location..."
              radius="md"
              size="md"
              className="pt-2"
            />
          </Collapse>
        </Stack>

        <Divider my="lg" />

        <Stack gap="md">
          <Group
            className="justify-between items-center cursor-pointer select-none"
            onClick={toggleTeams}
          >
            <Title order={4}>Teams</Title>
            <Button variant="light" size="xs">
              {teamsOpened ? 'Hide' : 'Show'}
            </Button>
          </Group>
          <Collapse in={teamsOpened}>
            <Stack gap="md" className="pt-2">
              {teams.map((team, idx) => (
                <Card key={team.id} withBorder radius="md" shadow="xs" p="md">
                  <Stack gap="xs">
                    <TextInput
                      label="Team Name"
                      value={team.name}
                      onChange={(e) =>
                        teamsHandlers.setItem(idx, {
                          ...team,
                          name: e.currentTarget.value,
                        })
                      }
                    />
                    <Textarea
                      label="Team Description"
                      minRows={2}
                      value={team.description}
                      onChange={(e) =>
                        teamsHandlers.setItem(idx, {
                          ...team,
                          description: e.currentTarget.value,
                        })
                      }
                    />
                  </Stack>
                </Card>
              ))}
              {canEdit && (
                <Button size="sm" variant="outline" onClick={addTeam}>
                  Add Team
                </Button>
              )}
            </Stack>
          </Collapse>
        </Stack>

        <Divider my="lg" />

        {canEdit && (
          <Group className="justify-end pt-6">
            <Button size="md" radius="md" bg="green" onClick={handleSave}>
              Save Changes
            </Button>
          </Group>
        )}
      </Paper>
    </Container>
  )
}

export default OrganizationSettings
