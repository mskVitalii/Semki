import type { SearchResult } from '@/common/types'
import {
  Anchor,
  Button,
  Group,
  Paper,
  Stack,
  Text,
  Title,
  Tooltip,
} from '@mantine/core'
import {
  IconBrandSlack,
  IconBrandTelegram,
  IconBrandWhatsapp,
  IconHash,
  IconMail,
  IconPhone,
} from '@tabler/icons-react'
import Interpretation from './Interpretation'

type UserResultCardProps = {
  data: SearchResult
}

function UserResultCard({ data }: UserResultCardProps) {
  const { user } = data
  const { contact } = user

  const contacts = [
    {
      key: 'telegram',
      value: contact.telegram,
      icon: IconBrandTelegram,
      href: `https://t.me/${contact.telegram.replace('@', '')}`,
    },
    {
      key: 'slack',
      value: contact.slack,
      icon: IconBrandSlack,
      href: `https://slack.com/app_redirect?channel=${contact.slack}`,
    },
    {
      key: 'email',
      value: contact.email,
      icon: IconMail,
      href: `mailto:${contact.email}`,
    },
    {
      key: 'telephone',
      value: contact.telephone,
      icon: IconPhone,
      href: `tel:${contact.telephone}`,
    },
    {
      key: 'whatsapp',
      value: contact.whatsapp,
      icon: IconBrandWhatsapp,
      href: `https://wa.me/${contact.whatsapp.replace('+', '')}`,
    },
  ].filter((c) => c.value)

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
        <Group justify="space-between" align="space-between" w={'100%'} grow>
          <Anchor component="a" href={`/profile/${data.user._id}`} mt={5}>
            <Title
              order={2}
              size="md"
              fw={600}
              className="leading-relaxed text-2xl! decoration-green-500!"
            >
              {data.user.name}
            </Title>
          </Anchor>
          <Group gap="xs" w={'min-content'} align="center" justify="flex-end">
            <IconHash className="w-3 h-3 text-slate-400" />
            <Text size="xs" c="dimmed" className="font-mono">
              {data.user._id}
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
        <Interpretation interpretation={data.description} />
        <Group gap="xs" mt="sm" wrap="wrap">
          {contacts.map(({ key, icon: Icon, href }) => (
            <Tooltip key={key} label={key}>
              <Button
                variant="light"
                size="xs"
                component="a"
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                leftSection={<Icon size={16} />}
              >
                {key}
              </Button>
            </Tooltip>
          ))}
        </Group>
      </Stack>
    </Paper>
  )
}

export default UserResultCard
