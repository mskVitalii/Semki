import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { useOrganizationStore } from '@/stores/organizationStore'
import { useUserStore } from '@/stores/userStore'
import { Avatar, Box, Text, UnstyledButton } from '@mantine/core'
import {
  IconLogout,
  IconPlus,
  IconSitemap,
  IconUser,
} from '@tabler/icons-react'
import { useMutation } from '@tanstack/react-query'
import { Link, useNavigate } from 'react-router-dom'
import History from './History'

interface SidebarProps {
  onNewChat: () => void
}

export function Sidebar({ onNewChat }: SidebarProps) {
  const navigate = useNavigate()
  const logout = useAuthStore((s) => s.logout)
  const claims = useAuthStore((s) => s.claims)
  const { organizationDomain } = useOrganizationStore()
  const refreshToken = useAuthStore((s) => s.refreshToken)
  const user = useUserStore((s) => s.user)
  const logoutMutation = useMutation({
    mutationFn: async () => {
      await api.post('/api/v1/logout', { refresh_token: refreshToken })
    },
    onSuccess: () => {
      logout()
      navigate('/login', { replace: true })
    },
  })

  return (
    <Box className="relative h-screen flex flex-col border-r-2! border-[var(--mantine-color-dark-6)]!">
      <div className="flex flex-col h-full p-6! flex-1 space-y-6!">
        <Link to={`/profile/${claims?._id}`} className="no-underline mb-4">
          <UnstyledButton className="w-full flex items-center gap-3 p-3 rounded-lg">
            <Avatar size="md" radius="xl">
              <IconUser size={20} />
            </Avatar>
            <Text size="lg" fw={500} c="green" className="hover:text-blue-100!">
              {user?.email ?? 'Profile'}
            </Text>
          </UnstyledButton>
        </Link>

        <Link to={`/organization`} className="no-underline mb-4">
          <UnstyledButton className="w-full flex items-center gap-3 p-3 rounded-lg">
            <Avatar size="md" radius="xl">
              <IconSitemap size={20} />
            </Avatar>
            <Text
              size="md"
              fw={500}
              c="green"
              className="hover:text-blue-100! capitalize"
            >
              {organizationDomain ?? 'Organization'}
            </Text>
          </UnstyledButton>
        </Link>

        <History />

        <UnstyledButton
          className="w-full flex items-center mb-0! mb justify-center gap-2 p-3! rounded-lg! bg-green-500 hover:bg-green-600! text-white!"
          onClick={onNewChat}
        >
          <IconPlus size={20} />
          <Text size="sm" fw={500}>
            New
          </Text>
        </UnstyledButton>

        <div className="border-t border-gray-200! p-4!">
          <UnstyledButton
            onClick={() => logoutMutation.mutate()}
            className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-100 text-red-600"
          >
            <IconLogout size={20} />
            <Text size="sm" fw={500}>
              Logout
            </Text>
          </UnstyledButton>
        </div>
      </div>
    </Box>
  )
}
