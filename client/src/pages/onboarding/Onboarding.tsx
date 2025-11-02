import { setPassword, type AuthErrorResponse } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import {
  Button,
  Paper,
  PasswordInput,
  Stack,
  Text,
  type PaperProps,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { notifications } from '@mantine/notifications'
import { useMutation } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'

export function Onboarding(props: PaperProps) {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setAuth } = useAuthStore()

  useEffect(() => {
    const accessTokenFromUrl = searchParams.get('accessToken')
    const refreshTokenFromUrl = searchParams.get('refreshToken')
    const errorFromUrl = searchParams.get('error')

    if (errorFromUrl) {
      notifications.show({
        title: 'Authentication Error',
        message: decodeURIComponent(errorFromUrl),
        color: 'red',
        autoClose: false,
      })
      navigate('/login', { replace: true })
      return
    }

    if (!accessTokenFromUrl || !refreshTokenFromUrl) {
      notifications.show({
        title: 'Invalid Link',
        message: 'This onboarding link is invalid or expired.',
        color: 'red',
      })
      navigate('/login', { replace: true })
      return
    }

    setAuth(accessTokenFromUrl, refreshTokenFromUrl)
  }, [navigate, searchParams, setAuth])

  const form = useForm({
    initialValues: {
      password: '',
      confirmPassword: '',
    },

    validate: {
      password: (val) =>
        val.length < 8 ? 'Password should include at least 8 characters' : null,
      confirmPassword: (val, values) =>
        val !== values.password ? 'Passwords do not match' : null,
    },
  })

  const setPasswordMutation = useMutation({
    mutationFn: setPassword,
    onSuccess: () => {
      notifications.show({
        title: 'Success',
        message: 'Password set successfully',
        color: 'green',
      })
      navigate('/chat', { replace: true })
    },
    onError: (error: AxiosError<AuthErrorResponse>) => {
      console.error(error)
      notifications.show({
        title: 'Error',
        message: error.response?.data?.message || 'Failed to set password',
        color: 'red',
      })
    },
  })

  const handleSubmit = (values: typeof form.values) => {
    setPasswordMutation.mutate({
      password: values.password,
    })
  }

  return (
    <div className="flex items-center justify-center min-h-screen max-w-screen w-full bg-gray-900">
      <Paper radius="md" p="lg" withBorder {...props}>
        <Text size="lg" fw={500} mb="md">
          Set Your Password
        </Text>

        <Text size="sm" c="dimmed" mb="xl">
          Welcome! Please create a secure password for your account.
        </Text>

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            <PasswordInput
              required
              label="Password"
              placeholder="Create a password"
              value={form.values.password}
              onChange={(event) =>
                form.setFieldValue('password', event.currentTarget.value)
              }
              error={
                form.errors.password &&
                'Password should include at least 6 characters'
              }
              radius="md"
            />

            <PasswordInput
              required
              label="Confirm Password"
              placeholder="Confirm your password"
              value={form.values.confirmPassword}
              onChange={(event) =>
                form.setFieldValue('confirmPassword', event.currentTarget.value)
              }
              error={form.errors.confirmPassword && 'Passwords do not match'}
              radius="md"
            />
          </Stack>

          <Button
            type="submit"
            bg="green"
            radius="xl"
            fullWidth
            mt="xl"
            loading={setPasswordMutation.isPending}
          >
            Set Password
          </Button>
        </form>
      </Paper>
    </div>
  )
}

export default Onboarding
