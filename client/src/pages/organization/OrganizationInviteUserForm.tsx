import type { InviteUserData } from '@/api/user'
import { OrganizationRoles, type Organization } from '@/common/types'
import { useOrganizationStore } from '@/stores/organizationStore'
import {
  Button,
  Card,
  Group,
  Loader,
  Select,
  Stack,
  Textarea,
  TextInput,
} from '@mantine/core'
import { useForm } from '@mantine/form'

const defaultUsers: InviteUserData[] = [
  {
    email: 'lara.croft@croftventures.com',
    name: 'Lara Croft',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Archaeologist and explorer specializing in ancient civilizations and lost artifacts. Operates independently with expertise in survival, languages, and historical research.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'shepard@n7-alliance.org',
    name: 'Commander Shepard',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Alliance Navy officer with N7 certification. Skilled in tactical command, diplomacy, and leading high-risk missions across interstellar space.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'gerald@kaer-morhen.wt',
    name: 'Geralt of Rivia',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Professional monster hunter trained at Kaer Morhen. Expert in alchemy, sword combat, and negotiation. Known for neutrality and precision in contract work.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'ripley@weyland-yutani.com',
    name: 'Ellen Ripley',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Warrant officer and xenobiology specialist. Experienced in crisis situations involving unknown organisms and corporate cover-ups.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'neo@zion.net',
    name: 'Thomas Anderson',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Former software engineer who discovered the truth of simulated reality. Possesses advanced digital manipulation and combat abilities.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'd.v@overwatch.io',
    name: 'Hana Song',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Professional gamer turned mech pilot. Uses advanced MEKA technology for defense operations and rapid-response missions.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'arthur.morgan@van.der.linde',
    name: 'Arthur Morgan',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Seasoned outlaw and gunslinger. Skilled in tracking, horseback combat, and wilderness survival. Known for loyalty and complex moral code.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'solid.snake@foxhound.org',
    name: 'Solid Snake',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Elite covert operative specializing in infiltration, espionage, and counter-terrorism. Former member of FOXHOUND unit.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'ciri@cintra.gov',
    name: 'Cirilla Fiona Elen Riannon',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `Princess of Cintra and bearer of the Elder Blood. Trained in swordsmanship and capable of traversing between worlds using ancient magic.`,
    },
    organizationRole: OrganizationRoles.USER,
  },
  {
    email: 'g-man@gmail.com',
    name: 'Gordon Freeman',
    semantic: {
      team: '',
      level: '',
      location: '',
      description: `SUBJECT:  FREEMAN
STATUS:  HIRED
AWAITING ASSIGNMENT

Physics Researcher
`,
    },
    organizationRole: OrganizationRoles.USER,
  },
]

const createDefaultUser = (
  organization: Organization | null,
): InviteUserData => {
  if (!organization) throw new Error('No organization')
  const location = organization.semantic.locations[0]?.id ?? ''
  const user = defaultUsers[Math.floor(Math.random() * defaultUsers.length)]
  if (user.semantic) user.semantic.location = location
  return user
}

type OrganizationInviteUserFormProps = {
  onSave: (newUser: InviteUserData) => void
}

function OrganizationInviteUserForm({
  onSave,
}: OrganizationInviteUserFormProps) {
  const organization = useOrganizationStore((s) => s.organization)

  const form = useForm<InviteUserData>({
    mode: 'uncontrolled',
    initialValues: createDefaultUser(organization),
    validate: {
      name: (value) => (!value ? 'Name is required' : null),
      email: (value) => {
        if (!value) return 'Email is required'
        if (!/^\S+@\S+\.\S+$/.test(value)) return 'Invalid email format'
        return null
      },
      semantic: {
        team: (value) => (!value ? 'Team is required' : null),
        level: (value) => (!value ? 'Level is required' : null),
      },
      organizationRole: (value) => (!value ? 'Role is required' : null),
    },
  })

  const handleReset = () => {
    const defaults = createDefaultUser(organization)
    form.setValues(defaults)
    form.setInitialValues(defaults)
    form.reset()
  }
  const handleSubmit = (values: InviteUserData) => {
    onSave(values)
    handleReset()
  }

  if (!organization)
    return (
      <div className="flex-1 flex items-center justify-center h-full">
        <Loader color="green" />
      </div>
    )

  return (
    <Card withBorder radius="md" shadow="xs" p="md">
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="sm">
          <Group grow>
            <TextInput
              label="Full Name"
              placeholder="Enter full name"
              withAsterisk
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('name')}
              {...form.getInputProps('name')}
            />
            <TextInput
              label="Email"
              placeholder="user@example.com"
              withAsterisk
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('email')}
              {...form.getInputProps('email')}
            />
          </Group>

          <Select
            label="Location"
            placeholder="Select location"
            withAsterisk
            allowDeselect={false}
            data={
              organization.semantic.locations
                .filter((l) => l.id !== undefined || l.id !== null)
                .map((l) => ({
                  value: l.id!,
                  label: l.name,
                })) ?? []
            }
            styles={{ label: { marginBottom: '0.75rem' } }}
            key={form.key('semantic.location')}
            {...form.getInputProps('semantic.location')}
          />

          <Group grow>
            <Select
              label="Team"
              placeholder="Select team"
              withAsterisk
              allowDeselect={false}
              data={
                organization.semantic.teams
                  .filter((t) => t.id !== undefined || t.id !== null)
                  .map((t) => ({
                    value: t.id!,
                    label: t.name,
                  })) ?? []
              }
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('semantic.team')}
              {...form.getInputProps('semantic.team')}
            />
            <Select
              label="Level"
              placeholder="Select level"
              withAsterisk
              allowDeselect={false}
              data={
                organization.semantic.levels
                  .filter((l) => l.id !== undefined || l.id !== null)
                  .map((l) => ({
                    value: l.id!,
                    label: l.name,
                  })) ?? []
              }
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('semantic.level')}
              {...form.getInputProps('semantic.level')}
            />
          </Group>

          <Textarea
            label="Description"
            placeholder="Enter user description"
            minRows={form.values.semantic?.description?.split('\n').length ?? 4}
            withAsterisk
            styles={{
              label: { marginBottom: '0.75rem' },
              input: { resize: 'vertical' },
            }}
            key={form.key('semantic.description')}
            {...form.getInputProps('semantic.description')}
          />

          <Group grow>
            <Select
              label="Role"
              placeholder="Select role"
              data={Object.values(OrganizationRoles).map((r) => ({
                value: r,
                label: r,
              }))}
              allowDeselect={false}
              styles={{ label: { marginBottom: '0.75rem' } }}
              key={form.key('organizationRole')}
              {...form.getInputProps('organizationRole')}
            />
          </Group>

          <Group justify="flex-end" pt="md">
            <Button variant="subtle" onClick={handleReset} type="button">
              Reset
            </Button>
            <Button radius="md" type="submit">
              Add User
            </Button>
          </Group>
        </Stack>
      </form>
    </Card>
  )
}

export default OrganizationInviteUserForm
