import { fetchChatById } from '@/api/chat'
import { useCreateChat } from '@/common/hooks/useCreateChat'
import { MainLayout } from '@/common/SidebarLayout'
import type {
  CreateChatResponse,
  GetChatResponse,
  SearchRequest,
  SearchResult,
} from '@/common/types'
import { useAuthStore } from '@/stores/authStore'
import { useOrganizationStore } from '@/stores/organizationStore'
import { Alert, Anchor, Badge, Card, Group, Stack, Title } from '@mantine/core'
import { useListState } from '@mantine/hooks'
import { IconAlertCircle } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import SearchForm from './SearchForm'
import UserResultCard from './UserResultCard'

const Chat: React.FC = () => {
  const [users, usersHandlers] = useListState<SearchResult>([])
  const access_token = useAuthStore((state) => state.accessToken)
  const [error, setError] = useState<string>('')
  const abortControllerRef = useRef<AbortController | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const { mutateAsync: createChat } = useCreateChat()
  const [req, setReq] = useState<SearchRequest>()

  const handleClear = useCallback((): void => {
    usersHandlers.setState([])
    setError('')
    setReq({
      q: '',
      teams: [],
      levels: [],
      locations: [],
      limit: 10,
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const { chatId } = useParams<{ chatId?: string }>()
  const navigate = useNavigate()

  const {
    data: chat,
    isError,
    error: chatLoadError,
  } = useQuery<GetChatResponse, AxiosError>({
    queryKey: ['chat', chatId],
    queryFn: () => fetchChatById(chatId!),
    enabled: !!chatId,
  })
  console.log('current chat ', chat, isError)

  useEffect(() => {
    if (!chat) return
    if (chat.messages.length === 0) return
    const chatResponse = chat.messages[0] as CreateChatResponse
    setReq({
      q: chatResponse.title,
      teams: chatResponse.teams ?? [],
      levels: chatResponse.levels ?? [],
      locations: chatResponse.locations ?? [],
      limit: chatResponse.limit ?? 10,
    })
    usersHandlers.setState(chat.messages.filter((x) => 'user' in x))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chat])

  useEffect(() => {
    if (!chatId) handleClear()
  }, [handleClear, chatId])

  const handleStream = useCallback(
    async (
      question: string,
      chatId: string,
      filters: Omit<SearchRequest, 'q' | 'limit'>,
    ) => {
      handleClear()
      setIsLoading(true)
      const controller = new AbortController()
      abortControllerRef.current = controller

      try {
        const encodedQuestion = encodeURIComponent(question)
        const url = `${import.meta.env.VITE_API_URL}/api/v1/search`

        const params = new URLSearchParams()
        params.append('q', encodedQuestion)
        params.append('chatId', chatId)

        Object.entries(filters || {}).forEach(([key, value]) => {
          if (Array.isArray(value)) {
            value.forEach((v) => params.append(key, String(v)))
          } else if (value !== undefined && value !== null) {
            params.append(key, String(value))
          }
        })

        const response = await fetch(`${url}?${params.toString()}`, {
          headers: {
            Authorization: `Bearer ${access_token}`,
          },
          signal: controller.signal,
        })

        if (!response.body) throw new Error('No response body')

        const stream = response.body
          .pipeThrough(new TextDecoderStream()) // превращает байты в строки
          .pipeThrough(
            new TransformStream({
              transform(chunk, controller) {
                chunk.split('\n\n').forEach((line) => controller.enqueue(line))
              },
            }),
          )

        const reader = stream.getReader()
        let buffer = ''
        while (true) {
          const { value, done } = await reader.read()
          if (done) break
          const line = value.replace(/^event:result\s*/, '').trim()
          if (!line || line === '[DONE]') continue

          buffer += line.replace(/^data:\s*/, '')

          try {
            const parsed = JSON.parse(buffer)
            usersHandlers.append(parsed)
            buffer = ''
          } catch (err) {
            console.warn('Failed to parse SSE:', line, err)
          }
        }
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          setError('Request was cancelled')
        } else if (err instanceof Error) {
          setError(err.message)
          console.error(err)
        } else {
          setError('Unknown error occurred')
        }
      } finally {
        setIsLoading(false)
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

  const handleSubmit = (
    question: string,
    filters?: Omit<SearchRequest, 'q' | 'limit'>,
  ): void => {
    // console.log(isLoading, !question.trim())
    if (isLoading || !question.trim()) return

    const effectiveFilters: Omit<SearchRequest, 'q' | 'limit'> = filters ?? {
      teams: [],
      levels: [],
      locations: [],
    }

    // Run async work without returning a Promise to match SearchForm's onSearch signature
    ;(async () => {
      try {
        // Chat
        const chat = await createChat({
          query: question.trim(),
          ...effectiveFilters,
        })

        // Search
        await handleStream(question.trim(), chat.id, effectiveFilters)
        navigate(`/chat/${chat.id}`, { replace: false })
      } catch (err) {
        console.error('handleSubmit error', err)
      }
    })()
  }

  const sortedUsers = useMemo(
    () => users.sort((a, b) => b.score - a.score),
    [users],
  )

  const organization = useOrganizationStore((s) => s.organization)

  const badges = useMemo(() => {
    if (!req || !organization) return []
    const { teams, levels, locations } = organization.semantic

    const selectedTeams =
      req.teams
        .map((id) => teams.find((t) => t.id === id)?.name)
        .filter(Boolean) ?? []
    const selectedLevels =
      req.levels
        .map((id) => levels.find((l) => l.id === id)?.name)
        .filter(Boolean) ?? []
    const selectedLocations =
      req.locations
        .map((id) => locations.find((l) => l.id === id)?.name)
        .filter(Boolean) ?? []

    return [
      ...selectedTeams.map((name) => ({ name, color: 'violet' })),
      ...selectedLevels.map((name) => ({ name, color: 'blue' })),
      ...selectedLocations.map((name) => ({ name, color: 'green' })),
    ]
  }, [req, organization])

  return (
    <MainLayout>
      <Card
        shadow="lg"
        padding="xl"
        radius="md"
        className={`min-h-screen ${sortedUsers.length === 0 ? 'h-full!' : ''} w-full p-4`}
      >
        <Stack className="h-full" justify="space-between">
          {/* Header */}
          <Group justify="space-between" align="center">
            <Anchor
              variant="gradient"
              gradient={{ from: 'green', to: 'white' }}
              fw={500}
              fz="lg"
              className="cursor-default"
            >
              <Title order={2}>🌻 Semki</Title>
            </Anchor>
          </Group>

          {/* Input Section */}
          {!chat && (
            <SearchForm
              onSearch={handleSubmit}
              onCancel={handleCancel}
              isLoading={isLoading}
              req={req}
            />
          )}

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

          {isError && chatLoadError && (
            <Alert
              icon={<IconAlertCircle size={16} />}
              title="Error"
              color="red"
              variant="light"
            >
              {chatLoadError.message || 'Failed to load chat.'}
            </Alert>
          )}

          {/* Response Display */}
          <Group justify="space-between" mb="sm">
            <Title order={1} fw={900} className="text-5xl" mt={'lg'}>
              {req?.q ?? 'Response'}
            </Title>
            <Group gap="xs" mt="xs" wrap="wrap">
              {badges.map((b) => (
                <Badge size="lg" key={b.name} color={b.color}>
                  {b.name}
                </Badge>
              ))}
            </Group>

            {isLoading && (
              <Badge color="blue" variant="dot" size="sm">
                Streaming...
              </Badge>
            )}
          </Group>

          {sortedUsers.length > 0 && (
            <div className="w-full mx-auto p-4">
              {/* Found Users List */}
              <div className="mb-4">
                <Stack gap="md">
                  {sortedUsers.map((userRes, i) => (
                    <UserResultCard
                      data={userRes}
                      key={userRes.user?._id ?? i}
                    />
                  ))}
                </Stack>
              </div>
            </div>
          )}

          {sortedUsers.length === 1 && !isLoading && (
            <div className="flex-1 flex items-center justify-center">
              <Title order={3} c="dimmed">
                No results found. Try adjusting your search criteria.
              </Title>
            </div>
          )}
        </Stack>
      </Card>
    </MainLayout>
  )
}

export default Chat
