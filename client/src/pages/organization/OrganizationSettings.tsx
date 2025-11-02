import { useAuthStore } from '@/stores/authStore'
import { useOrganizationStore } from '@/stores/organizationStore'
import {
  Container,
  Divider,
  Group,
  Paper,
  Stack,
  TextInput,
  Title,
} from '@mantine/core'
import { LevelSection } from './LevelsSection'
import { LocationSection } from './LocationSection'
import { TeamSection } from './TeamSection'
import { TitleSection } from './TitleSection'

export function OrganizationSettings() {
  const organization = useOrganizationStore((s) => s.organization)
  const isAdmin = useAuthStore((s) => s.isAdmin)

  if (!organization) return <div>Loading...</div>

  return (
    <Container className="py-12!">
      <Paper className="p-8! max-w-3xl! mx-auto! space-y-8!">
        <Group className="justify-between items-center">
          <Title order={2}>Organization Settings</Title>
        </Group>

        <Stack gap="md">
          <TitleSection disabled={!isAdmin} />
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

        <LevelSection disabled={!isAdmin} />

        <Divider my="lg" />

        <LocationSection disabled={!isAdmin} />

        <Divider my="lg" />

        <TeamSection disabled={!isAdmin} />

        <Divider my="lg" />
      </Paper>
    </Container>
  )
}

export default OrganizationSettings
