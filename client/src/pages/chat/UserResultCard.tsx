import { UserStatuses, type SearchResult } from '@/common/types'
import { useOrganizationStore } from '@/stores/organizationStore'
import {
  Anchor,
  Badge,
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
  IconMapPin,
  IconPhone,
} from '@tabler/icons-react'
import React, { useMemo } from 'react'
import Interpretation from './Interpretation'

type UserResultCardProps = {
  data: SearchResult
}

function UserResultCard({ data }: UserResultCardProps) {
  const { user } = data
  const { contact } = user
  const organization = useOrganizationStore((s) => s.organization)
  const level = useMemo(
    () =>
      organization?.semantic.levels.find((l) => l.id === user.semantic.level)
        ?.name ?? user.semantic.level,
    [organization?.semantic.levels, user.semantic.level],
  )
  const location = useMemo(
    () =>
      organization?.semantic.locations.find(
        (l) => l.id === user.semantic.location,
      )?.name ?? user.semantic.location,
    [organization?.semantic.locations, user.semantic.location],
  )

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
      value: contact.email === '' ? user.email : contact.email,
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
        <Text size="sm" className="leading-relaxed text-slate-700">
          {user.semantic?.description}
        </Text>

        <Group gap="xs" mt="xs">
          {user.status == UserStatuses.DELETED && (
            <Badge color="red">{UserStatuses.DELETED}</Badge>
          )}
          {level && <Badge color="blue">{level}</Badge>}
          {location && (
            <Badge leftSection={<IconMapPin size={12} />} color="green">
              {location}
            </Badge>
          )}
        </Group>

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

export default React.memo(UserResultCard)
