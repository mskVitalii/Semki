import { chatHistory } from '@/api/chat'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { useOrganizationStore } from '@/stores/organizationStore'
import { useUserStore } from '@/stores/userStore'
import { Avatar, Box, ScrollArea, Text, UnstyledButton } from '@mantine/core'
import { useIntersection } from '@mantine/hooks'
import {
  IconLogout,
  IconPlus,
  IconSitemap,
  IconUser,
} from '@tabler/icons-react'
import { useInfiniteQuery, useMutation } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'

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

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteQuery({
      queryKey: ['chatHistory'],
      queryFn: chatHistory,
      getNextPageParam: (lastPage) => lastPage.nextCursor,
      initialPageParam: undefined as string | undefined,
    })

  const { ref, entry } = useIntersection({ threshold: 1 })

  useEffect(() => {
    if (entry?.isIntersecting && hasNextPage && !isFetchingNextPage) {
      fetchNextPage()
    }
  }, [entry?.isIntersecting, hasNextPage, isFetchingNextPage, fetchNextPage])

  const allChats = data?.pages.flatMap((page) => page.chats) ?? []

  return (
    <Box className="relative h-screen flex flex-col border-r-2! border-[var(--mantine-color-dark-6)]!">
      <div className="flex flex-col h-full p-6! flex-1 space-y-6!">
        <Link to={`/profile/${claims?._id}`} className="no-underline mb-4">
          <UnstyledButton className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-100">
            <Avatar size="md" radius="xl">
              <IconUser size={20} />
            </Avatar>
            <Text size="lg" fw={500} c="green">
              {user?.email ?? 'Profile'}
            </Text>
          </UnstyledButton>
        </Link>

        <Link to={`/organization`} className="no-underline mb-4">
          <UnstyledButton className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-100">
            <Avatar size="md" radius="xl">
              <IconSitemap size={20} />
            </Avatar>
            <Text size="md" fw={500} c="green">
              {organizationDomain ?? 'Organization'}
            </Text>
          </UnstyledButton>
        </Link>

        <Text size="xs" fw={600} className="text-gray-600 mb-2 px-2">
          History
        </Text>

        <ScrollArea className="flex-1">
          <div className="space-y-1">
            {allChats.map((chat) => (
              <UnstyledButton
                key={chat.id}
                className="w-full p-3 rounded-lg hover:bg-gray-100 text-left"
                onClick={() => navigate(`/chat/${chat.id}`, { replace: true })}
              >
                <Text size="sm" className="text-gray-800 truncate">
                  {chat.title}
                </Text>
                <Text size="xs" className="text-gray-500 mt-1">
                  {new Date(chat.createdAt).toLocaleDateString('ru-RU')}
                </Text>
              </UnstyledButton>
            ))}
            {hasNextPage && (
              <div ref={ref} className="py-2 text-center">
                {isFetchingNextPage && (
                  <Text size="xs" className="text-gray-500">
                    Loading...
                  </Text>
                )}
              </div>
            )}
          </div>
        </ScrollArea>

        <UnstyledButton
          className="w-full flex items-center justify-center gap-2 p-3 mb-4 rounded-lg bg-blue-500 hover:bg-blue-600 text-white"
          onClick={onNewChat}
        >
          <IconPlus size={20} />
          <Text size="sm" fw={500}>
            New
          </Text>
        </UnstyledButton>
      </div>

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
    </Box>
  )
}
