import { type SearchResult } from '@/common/types'
import UserBadges from '@/common/UserBadges'
import UserContacts from '@/common/UserContacts'
import { Anchor, Group, Paper, Stack, Text, Title } from '@mantine/core'
import { IconHash } from '@tabler/icons-react'
import React from 'react'
import Interpretation from './Interpretation'

type UserResultCardProps = {
  data: SearchResult
}

function UserResultCard({ data }: UserResultCardProps) {
  const { user } = data

  return (
    <Paper
      key={user._id}
      p="lg"
      radius="md"
      className="border border-slate-200 bg-slate-50"
      withBorder
    >
      <Stack gap="sm">
        {/* Question Header */}
        <Group justify="space-between" align="space-between" w={'100%'} grow>
          <Anchor component="a" href={`/profile/${user._id}`} mt={5}>
            <Title
              order={2}
              size="md"
              fw={600}
              className="leading-relaxed text-2xl! decoration-green-500!"
            >
              {user.name}
            </Title>
          </Anchor>
          <Group gap="xs" w={'min-content'} align="center" justify="flex-end">
            <IconHash className="w-3 h-3 text-slate-400" />
            <Text size="xs" c="dimmed" className="font-mono">
              {user._id}
            </Text>
          </Group>
        </Group>

        {/* Reason */}
        <Text size="sm" c="dimmed" className="text-slate-500 leading-relaxed">
          <span className="font-semibold">🔥 Hot:</span>
          <Anchor
            ml="md"
            className="cursor-default!"
            variant="gradient"
            gradient={{ from: 'yellow', to: 'red' }}
          >
            {data.score}
          </Anchor>
        </Text>
        <Text size="sm" className="leading-relaxed text-slate-700">
          {user.semantic?.description}
        </Text>

        <UserBadges user={user} />

        <Interpretation interpretation={data.description} />
        <UserContacts contact={user.contact} />
      </Stack>
    </Paper>
  )
}

export default React.memo(UserResultCard)
