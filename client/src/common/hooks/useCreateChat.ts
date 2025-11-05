import { api } from '@/api/client'
import { type CreateChatRequest, type CreateChatResponse } from '@/common/types'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useRef } from 'react'

export const useCreateChat = () => {
  const queryClient = useQueryClient()
  const abortControllerRef = useRef<AbortController | null>(null)

  const mutation = useMutation<CreateChatResponse, Error, CreateChatRequest>({
    mutationFn: async (data) => {
      const controller = new AbortController()
      abortControllerRef.current = controller

      const response = await api.post<CreateChatResponse>(
        '/api/v1/chat',
        data,
        {
          signal: controller.signal,
        },
      )

      return response.data
    },
    onSuccess: (data) => {
      console.log('useCreateChat', data)
      queryClient.invalidateQueries({
        queryKey: ['chatHistory'],
      })
    },
    onError: (err) => {
      console.error(err)
    },
  })

  const cancel = () => {
    if (!abortControllerRef.current) return
    abortControllerRef.current.abort()
    abortControllerRef.current = null
  }

  return { ...mutation, cancel }
}
