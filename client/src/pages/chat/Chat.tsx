import { MainLayout } from '@/common/SidebarLayout'
import type { SearchResult } from '@/common/types'
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
import React, { useCallback, useRef, useState } from 'react'
import SearchForm from './SearchForm'
import UserResultCard from './UserResultCard'

const Chat: React.FC = () => {
  const [users, usersHandlers] = useListState<SearchResult>([])
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [error, setError] = useState<string>('')
  const abortControllerRef = useRef<AbortController | null>(null)

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

  // TODO: request to get chat
  // TODO: button to retry if no answer in chat & no generation in progress

  const handleStream = useCallback(
    async (question: string) => {
      handleClear()
      setIsLoading(true)

      let eventSource: EventSource | null = null
      let isManualClose = false

      try {
        const encodedQuestion = encodeURIComponent(question)
        console.log(`request to ${import.meta.env.VITE_API_URL}/api/v1/search`)
        const endpoint = `${import.meta.env.VITE_API_URL}/api/v1/search`
        const url = `${endpoint}?question=${encodedQuestion}`

        eventSource = new EventSource(url)

        eventSource.onopen = () => {
          console.log('EventSource connection opened')
        }

        eventSource.onmessage = (event) => {
          if (event.data === '[DONE]' || event.data.includes('[DONE]')) {
            console.log('Stream completed successfully')
            eventSource?.close()
            setIsLoading(false)
            return
          }

          const parsed = parseSSELine(event.data)

          if (parsed) {
            usersHandlers.append(parsed)
          }
        }

        eventSource.onerror = (error) => {
          console.error('EventSource error:', error)
          if (isManualClose) eventSource?.close()

          if (eventSource?.readyState === EventSource.CLOSED) {
            console.log('Connection closed normally')
          } else if (eventSource?.readyState === EventSource.CONNECTING) {
            setError('Failed to establish connection')
          } else {
            setError('Connection error occurred')
          }

          setIsLoading(false)
          eventSource?.close()
        }

        abortControllerRef.current = {
          abort: () => {
            isManualClose = true
            eventSource?.close()
            setError('Request was cancelled')
            setIsLoading(false)
          },
        } as AbortController
      } catch (err) {
        if (err instanceof Error) {
          setError(`Error: ${err.message}`)
          console.error('Stream error:', err)
        } else {
          setError('An unknown error occurred')
        }
        setIsLoading(false)
      }

      return () => {
        isManualClose = true
        eventSource?.close()
        setIsLoading(false)
        abortControllerRef.current = null
      }
    },
    [handleClear, usersHandlers],
  )

  const handleCancel = useCallback((): void => {
    if (!abortControllerRef.current) return
    abortControllerRef.current.abort()
    abortControllerRef.current = null
  }, [])

  const handleSubmit = (question: string): void => {
    // TODO: create chat

    if (!isLoading && question.trim()) {
      handleStream(question.trim())
      console.log(question)
    }
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
