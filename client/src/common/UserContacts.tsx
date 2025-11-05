import { Button, Group, Tooltip } from '@mantine/core'
import {
  IconBrandSlack,
  IconBrandTelegram,
  IconBrandWhatsapp,
  IconMail,
  IconPhone,
} from '@tabler/icons-react'
import type { User } from './types'
import React from 'react'

function UserContacts({ contact }: { contact: User['contact'] }) {
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
  )
}

export default React.memo(UserContacts)
