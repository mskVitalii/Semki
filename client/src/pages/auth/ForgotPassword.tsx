import { requestPasswordReset } from '@/api/auth'
import {
  Anchor,
  Box,
  Button,
  Center,
  Container,
  Group,
  Paper,
  Text,
  TextInput,
  Title,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { notifications } from '@mantine/notifications'
import { IconArrowLeft } from '@tabler/icons-react'
import { useMutation } from '@tanstack/react-query'

export function ForgotPassword() {
  const form = useForm({
    initialValues: {
      email: import.meta.env.VITE_TEST_EMAIL ?? '',
    },
    validate: {
      email: (value) => (/^\S+@\S+\.\S+$/.test(value) ? null : 'Invalid email'),
    },
  })

  const mutation = useMutation({
    mutationFn: requestPasswordReset,
    onSuccess: () => {
      notifications.show({
        title: 'Success',
        message: 'Reset link sent to your email',
        color: 'green',
      })
      form.reset()
    },
    onError: (err) => {
      notifications.show({
        title: 'Error',
        message: err.message || 'Failed to send reset link',
        color: 'red',
      })
    },
  })

  const handleSubmit = (values: typeof form.values) => {
    mutation.mutate(values.email)
  }

  return (
    <div className="flex items-center justify-center min-h-screen max-w-screen w-screen bg-gray-900">
      <Container size={460} my={30}>
        <Title
          className="text-center text-2xl font-medium font-sans"
          ta="center"
        >
          Forgot your password?
        </Title>
        <Text c="dimmed" fz="sm" ta="center">
          Enter your email to get a reset link
        </Text>

        <Paper withBorder shadow="md" p={30} radius="md" mt="xl">
          <form onSubmit={form.onSubmit(handleSubmit)}>
            <TextInput
              label="Your email"
              placeholder="example@gmail.com"
              styles={{ label: { marginBottom: '0.75rem' } }}
              required
              {...form.getInputProps('email')}
            />
            <Group
              justify="space-between"
              mt="lg"
              className="flex-col-reverse sm:flex-row gap-3"
            >
              <Anchor
                c="dimmed"
                size="sm"
                href="/login"
                className="flex items-center justify-center sm:justify-start gap-1"
              >
                <Center inline>
                  <IconArrowLeft size={12} stroke={1.5} />
                  <Box ml={5}>Back to the login page</Box>
                </Center>
              </Anchor>
              <Button
                type="submit"
                loading={mutation.isPending}
                bg="green"
                className="bg-blue-500 text-white px-4 py-2 rounded w-full sm:w-auto"
              >
                Reset password
              </Button>
            </Group>
          </form>
        </Paper>
      </Container>
    </div>
  )
}

export default ForgotPassword
