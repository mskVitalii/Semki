import { MainLayout } from '@/common/SidebarLayout'
import type { User } from '@/common/types'
import UserBadges from '@/common/UserBadges'
import UserContacts from '@/common/UserContacts'
import { useOrganizationStore } from '@/stores/organizationStore'
import { Card, Divider, Group, Loader, Stack, Text, Title } from '@mantine/core'

function ProfileReadOnly({ user }: { user: User }) {
  const organization = useOrganizationStore((s) => s.organization)

  if (!organization)
    return (
      <MainLayout>
        <div className="flex items-center h-full">
          <Loader color="green" />
        </div>
      </MainLayout>
    )

  return (
    <MainLayout>
      <Card
        shadow="lg"
        padding="xl"
        radius="md"
        className="min-h-screen w-full p-4 overflow-auto!"
      >
        <Stack>
          <Group justify="space-between" mt="md">
            <Title order={2}>{user.name}</Title>
          </Group>
          <UserBadges user={user} />

          <Divider label="General" m="lg" />
          <Text>{user.email}</Text>
          <Text>{user.name}</Text>
          <Divider label="Semantic" m="lg" />
          <Text>
            {user.semantic?.description || 'No description provided.'}
          </Text>

          <Divider label="Contact" m="lg" />
          <UserContacts contact={user.contact} />
        </Stack>
      </Card>
    </MainLayout>
  )
}

export default ProfileReadOnly
