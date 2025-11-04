import type { GetUserHistoryResponse } from '@/common/types'
import { api } from './client'

export const chatHistory = async ({
  pageParam,
}: {
  pageParam: string | undefined
}) => {
  const response = await api.get<GetUserHistoryResponse>(
    '/api/v1/chat/history',
    {
      params: { cursor: pageParam },
    },
  )
  return response.data
}


export const fetchChatById = async (id: string) => {
  const res = await api.get(`/api/v1/chat/${id}`)
  return res.data
}