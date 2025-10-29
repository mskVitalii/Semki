import { type Level, type Organization, type Team } from '@/common/types'
import { useAuthStore } from '@/stores/authStore'
import { useOrganizationStore } from '@/stores/organizationStore'
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
import { notifications } from '@mantine/notifications'
import { useState } from 'react'
import { v4 as uuid } from 'uuid'

export function OrganizationSettings() {
  const remoteOrg = useOrganizationStore((s) => s.organization)
  const setRemoteOrg = useOrganizationStore((s) => s.setOrganization)
  const isAdmin = useAuthStore((s) => s.isAdmin)

  const [organization, setOrganization] = useState<Organization | null>(
    remoteOrg,
  )

  // TODO: change on Form
  const [title, setTitle] = useState(remoteOrg?.title ?? '')
  const [levels, levelsHandlers] = useListState<Level>(
    remoteOrg?.semantic.levels ?? [],
  )
  const [locations, locationsHandlers] = useListState<string>(
    remoteOrg?.semantic.locations ?? [],
  )
  const [teams, teamsHandlers] = useListState<Team>(
    remoteOrg?.semantic.teams ?? [],
  )

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

  const handleSave = () => {
    // TODO: patch API
    if (!organization) return
    const newOrg: Organization = {
      ...organization,
      title,
      semantic: {
        ...organization.semantic,
        levels,
        locations,
        teams,
      },
    }
    setOrganization(newOrg)
    setRemoteOrg(newOrg)
    notifications.show({ title: 'Success', message: 'Saved', color: 'green' })
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
            disabled={!isAdmin}
            onChange={(e) => setTitle(e.currentTarget.value)}
            radius="md"
            size="md"
            styles={{ label: { marginBottom: '0.75rem' } }}
          />
          <TextInput
            label="Current Plan"
            value={organization.plan}
            disabled
            radius="md"
            size="md"
            styles={{ label: { marginBottom: '0.75rem' } }}
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
                      disabled={!isAdmin}
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
                      disabled={!isAdmin}
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
              {isAdmin && (
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
              disabled={!isAdmin}
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
                      disabled={!isAdmin}
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
                      disabled={!isAdmin}
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
              {isAdmin && (
                <Button size="sm" variant="outline" onClick={addTeam}>
                  Add Team
                </Button>
              )}
            </Stack>
          </Collapse>
        </Stack>

        <Divider my="lg" />

        {isAdmin && (
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
