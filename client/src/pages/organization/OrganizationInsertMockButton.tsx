import { api } from '@/api/client'
import { Button } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { useMutation, useQueryClient } from '@tanstack/react-query'

function OrganizationInsertMockButton() {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async () => await api.post('/api/v1/organization/insert-mock'),
    onSuccess: async () => {
      notifications.show({
        title: 'Success',
        message: 'Mock data inserted successfully',
        color: 'green',
      })
      await queryClient.invalidateQueries({
        queryKey: ['organization'],
      })
      await queryClient.invalidateQueries({
        queryKey: ['organizationUsers'],
      })
    },
    onError: () => {
      notifications.show({
        title: 'Error',
        message: 'Failed to insert mock data',
        color: 'red',
      })
    },
  })

  const handleInsertMockData = () => mutation.mutate()

  return (
    <Button
      onClick={handleInsertMockData}
      className="grow-0!"
      w="min-content"
      variant="gradient"
      gradient={{ from: 'indigo', to: 'purple' }}
      loading={mutation.isPending}
    >
      Insert Mock ✨🦄
    </Button>
  )
}

export default OrganizationInsertMockButton
