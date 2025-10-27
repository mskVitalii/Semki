import { api } from './client'

export interface ChatHistory {
  id: string
  title: string
  createdAt: string
}

export interface HistoryResponse {
  chats: ChatHistory[]
  nextCursor?: string
}

export const chatHistory = async ({
  pageParam,
}: {
  pageParam: string | undefined
}) => {
  const response = await api.get<HistoryResponse>('/api/v1/chat/history', {
    params: { cursor: pageParam },
  })
  return response.data
}
