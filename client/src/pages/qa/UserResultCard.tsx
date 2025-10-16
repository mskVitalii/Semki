import type { SearchResult } from '@/common/types'
import { Badge, Group, Paper, Stack, Text } from '@mantine/core'
import { IconHash } from '@tabler/icons-react'
import Interpretation from './Interpretation'

type UserResultCardProps = {
  data: SearchResult
}

function UserResultCard({ data }: UserResultCardProps) {
  return (
    <Paper
      key={data.user._id}
      p="lg"
      radius="md"
      className="border border-slate-200 bg-slate-50"
      withBorder
    >
      <Stack gap="sm">
        {/* Question Header */}
        <Group justify="space-between" align="flex-start">
          <Badge size="sm" variant="light" color="blue">
            User {data.user._id}
          </Badge>

          <Group gap="xs">
            <IconHash className="w-3 h-3 text-slate-400" />
            <Text size="xs" c="dimmed" className="font-mono">
              {data.user._id}
            </Text>
          </Group>
        </Group>
        <Text size="md" fw={600} className="text-slate-800 leading-relaxed">
          {data.user.name}
        </Text>
        {/* Reason */}
        <div className="mt-2 pl-4 border-l-2 border-slate-300 ">
          <Text size="sm" c="dimmed" className="text-slate-500 leading-relaxed">
            <span className="font-semibold">Reasoning:</span> {data.score}
          </Text>
        </div>
        <Interpretation interpretation={data.description} />
      </Stack>
    </Paper>
  )
}

export default UserResultCard
