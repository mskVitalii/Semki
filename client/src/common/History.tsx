import { chatHistory } from '@/api/chat'
import { Card, ScrollArea, Text } from '@mantine/core'
import { useIntersection } from '@mantine/hooks'
import { useInfiniteQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'

export default function History() {
  const navigate = useNavigate()

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

  const grouped = allChats.reduce<Record<string, typeof allChats>>(
    (acc, chat) => {
      const date = new Date(chat.created_at * 1000).toLocaleDateString(
        'en-EN',
        {
          day: '2-digit',
          month: 'long',
          year: 'numeric',
        },
      )
      if (!acc[date]) acc[date] = []
      acc[date].push(chat)
      return acc
    },
    {},
  )
  return (
    <ScrollArea className="flex-1">
      <div className="space-y-6 px-3 pb-4">
        {Object.entries(grouped)
          .sort(([a], [b]) => new Date(b).getTime() - new Date(a).getTime())
          .map(([date, chats]) => (
            <div key={date}>
              <Text
                fw={600}
                size="sm"
                mb="lg"
                mt="lg"
                className="text-gray-400 uppercase tracking-wide mb-3 px-1"
              >
                {date}
              </Text>

              <div className="grid gap-3">
                {chats
                  .sort((a, b) => b.created_at - a.created_at)
                  .map((chat) => (
                    <Card
                      key={chat.id}
                      shadow="sm"
                      padding="md"
                      radius="md"
                      withBorder
                      className="cursor-pointer bg-gray-800/40 hover:bg-gray-700/60! transition-all duration-150"
                      onClick={() =>
                        navigate(`/chat/${chat.id}`, { replace: true })
                      }
                    >
                      <Text
                        size="sm"
                        fw={500}
                        className="text-gray-50 mb-1 truncate"
                      >
                        {chat.title}
                      </Text>
                      <Text size="xs" className="text-gray-400">
                        {new Date(chat.created_at * 1000).toLocaleTimeString(
                          'ru-RU',
                          {
                            hour: '2-digit',
                            minute: '2-digit',
                          },
                        )}
                      </Text>
                    </Card>
                  ))}
              </div>
            </div>
          ))}

        {hasNextPage && (
          <div ref={ref} className="pt-3 text-center">
            {isFetchingNextPage && (
              <Text size="xs" className="text-gray-500">
                Loading...
              </Text>
            )}
          </div>
        )}
      </div>
    </ScrollArea>
  )
}
