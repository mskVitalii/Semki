import { fetchChatById } from '@/api/chat'
import { useCreateChat } from '@/common/hooks/useCreateChat'
import { MainLayout } from '@/common/SidebarLayout'
import type { SearchResult } from '@/common/types'
import { useAuthStore } from '@/stores/authStore'
import {
  ActionIcon,
  Alert,
  Anchor,
  Badge,
  Card,
  Divider,
  Group,
  Stack,
  Text,
  Title,
  Tooltip,
} from '@mantine/core'
import { useListState } from '@mantine/hooks'
import { IconAlertCircle, IconRefresh } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import React, { useCallback, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import SearchForm from './SearchForm'
import UserResultCard from './UserResultCard'

// TODO: button to retry if no answer in chat & no generation in progress

// TODO: request to start new chat

const Chat: React.FC = () => {
  const [users, usersHandlers] = useListState<SearchResult>([])
  const access_token = useAuthStore((state) => state.accessToken)
  const [error, setError] = useState<string>('')
  const abortControllerRef = useRef<AbortController | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const { mutateAsync: createChat } = useCreateChat()

  const parseSSELine = (line: string): SearchResult | null => {
    try {
      if (line === '[DONE]') return null

      return JSON.parse(line)
    } catch {
      // If not JSON, treat as plain text
      console.error('Failed to parse:', line)
      return null
    }
  }

  const handleClear = useCallback((): void => {
    usersHandlers.setState([])
    setError('')
  }, [usersHandlers])

  const { chatId } = useParams<{ chatId?: string }>()
  // console.log('chatId', chatId)
  const { data: chat, isError } = useQuery({
    queryKey: ['chat', chatId],
    queryFn: () => fetchChatById(chatId!),
    enabled: !!chatId,
  })
  console.log('current chat ', chat, isError)

  const handleStream = useCallback(
    async (question: string, chatId: string) => {
      handleClear()
      setIsLoading(true)
      const controller = new AbortController()
      abortControllerRef.current = controller

      try {
        const encodedQuestion = encodeURIComponent(question)
        const url = `${import.meta.env.VITE_API_URL}/api/v1/search?question=${encodedQuestion}`
        const response = await fetch(
          url +
            new URLSearchParams({
              q: encodedQuestion,
              chatId,
            }).toString(),
          {
            headers: {
              Authorization: `Bearer ${access_token}`,
            },
            signal: controller.signal,
          },
        )

        if (!response.body) throw new Error('No response body')

        const reader = response.body.getReader()
        const decoder = new TextDecoder()

        let done = false
        while (!done) {
          const { value, done: readerDone } = await reader.read()
          done = readerDone
          if (value) {
            const chunk = decoder.decode(value, { stream: true })
            chunk.split('\n').forEach((line) => {
              if (line.trim() === '[DONE]') {
                done = true
                return
              }
              const parsed = parseSSELine(line)
              if (parsed) usersHandlers.append(parsed)
            })
          }
        }

        setIsLoading(false)
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          setError('Request was cancelled')
        } else if (err instanceof Error) {
          setError(err.message)
          console.error(err)
        } else {
          setError('Unknown error occurred')
        }
        setIsLoading(false)
      } finally {
        abortControllerRef.current = null
      }
    },
    [access_token, handleClear, usersHandlers],
  )

  const handleCancel = useCallback((): void => {
    if (!abortControllerRef.current) return
    abortControllerRef.current.abort()
    abortControllerRef.current = null
  }, [])

  const handleSubmit = async (question: string): Promise<void> => {
    console.log(isLoading, !question.trim())
    if (isLoading || !question.trim()) return

    // Chat
    const chat = await createChat({ message: question.trim() })
    console.log('handleSubmit chat', chat)

    // Search
    await handleStream(question.trim(), chat.id)
  }

  return (
    <MainLayout>
      <Card
        shadow="lg"
        padding="xl"
        radius="md"
        className="min-h-screen w-full p-4"
      >
        <Stack>
          {/* Header */}
          <Group justify="space-between" align="center">
            <Anchor
              variant="gradient"
              gradient={{ from: 'green', to: 'white' }}
              fw={500}
              fz="lg"
              href="#text-props"
            >
              <Title order={2}>🌻 Semki</Title>
            </Anchor>
          </Group>

          {/* Input Section */}
          <SearchForm
            onSearch={handleSubmit}
            onCancel={handleCancel}
            isLoading={isLoading}
          />

          {/* Error Display */}
          {error && (
            <Alert
              icon={<IconAlertCircle size={16} />}
              title="Error"
              color="red"
              variant="light"
              withCloseButton
              onClose={() => setError('')}
            >
              {error}
            </Alert>
          )}

          {/* Response Display */}
          {users.length > 0 && (
            <>
              <Group justify="space-between" mb="sm">
                <Group>
                  <Text fw={600} size="lg" className="text-gray-700">
                    Response
                  </Text>
                  {isLoading && (
                    <Badge color="blue" variant="dot" size="sm">
                      Streaming...
                    </Badge>
                  )}
                </Group>
                <Tooltip label="Clear response">
                  <ActionIcon
                    onClick={handleClear}
                    variant="subtle"
                    color="gray"
                  >
                    <IconRefresh size={18} />
                  </ActionIcon>
                </Tooltip>
              </Group>

              <div className="w-full mx-auto p-4">
                <Card
                  shadow="md"
                  radius="lg"
                  className="bg-white dark:bg-slate-900"
                  withBorder
                >
                  <Divider my="xl" />

                  {/* Found Users List */}
                  <div className="mb-4">
                    <Text fw={500} size="lg" className="text-slate-700  mb-4">
                      Found Users
                    </Text>

                    <Stack gap="md">
                      {users.map((userRes) => (
                        <UserResultCard data={userRes} key={userRes.user._id} />
                      ))}
                    </Stack>
                  </div>
                </Card>
              </div>
            </>
          )}
        </Stack>
      </Card>
    </MainLayout>
  )
}

export default Chat
