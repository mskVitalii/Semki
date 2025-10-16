import { login, type AuthErrorResponse } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import { useOrganizationStore } from '@/stores/organizationStore'
import {
  Anchor,
  Button,
  Checkbox,
  Divider,
  Group,
  Paper,
  PasswordInput,
  Stack,
  Text,
  TextInput,
  type PaperProps,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { upperFirst, useToggle } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'
import { useMutation } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { GoogleButton } from './GoogleIcon'

export function Auth(props: PaperProps) {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { accessToken, setAuth } = useAuthStore()
  const [type, toggle] = useToggle(['login', 'register'])
  const { organizationDomain } = useOrganizationStore()

  useEffect(() => {
    const accessTokenFromUrl = searchParams.get('accessToken')
    const refreshTokenFromUrl = searchParams.get('refresh')
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

    if (accessTokenFromUrl && refreshTokenFromUrl) {
      setAuth(accessTokenFromUrl, refreshTokenFromUrl)
      navigate('/qa', { replace: true })
      return
    }

    if (accessToken) {
      navigate('/qa', { replace: true })
    }
  }, [accessToken, navigate, searchParams, setAuth])

  const form = useForm({
    initialValues: {
      email: import.meta.env.VITE_TEST_EMAIL ?? '',
      name: '',
      password: import.meta.env.VITE_TEST_PASSWORD ?? '',
      terms: true,
    },

    validate: {
      email: (val) => (/^\S+@\S+$/.test(val) ? null : 'Invalid email'),
      password: (val) =>
        val.length <= 6
          ? 'Password should include at least 6 characters'
          : null,
    },
  })

  const loginMutation = useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      setAuth(data.access_token, data.refresh_token)
      navigate('/qa', { replace: true })
    },
    onError: (error: AxiosError<AuthErrorResponse>) => {
      console.error(error)
      notifications.show({
        title: 'Auth error',
        message: error.response?.data?.message || 'Wrong email or password',
        color: 'red',
      })
    },
  })

  const handleSubmit = (values: typeof form.values) => {
    if (type === 'login') {
      loginMutation.mutate({
        email: values.email,
        password: values.password,
        organization: organizationDomain,
      })
    } else {
      console.log('Register:', values)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen max-w-screen w-full bg-gray-900">
      <Paper radius="md" p="lg" withBorder {...props}>
        <Text size="lg" fw={500} className="first-letter:uppercase">
          {organizationDomain}'s Semki, {type} with
        </Text>

        <Group grow mb="md" mt="md">
          <Anchor
            component="a"
            href={`${import.meta.env.VITE_API_URL}/api/v1/google/login`}
          >
            <GoogleButton radius="xl">Google</GoogleButton>
          </Anchor>
          <Button variant="default" radius="xl" disabled>
            SSO
          </Button>
        </Group>

        <Divider
          label="Or continue with email"
          labelPosition="center"
          my="lg"
        />

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            {type === 'register' && (
              <TextInput
                label="Name"
                placeholder="Your name"
                value={form.values.name}
                onChange={(event) =>
                  form.setFieldValue('name', event.currentTarget.value)
                }
                radius="md"
              />
            )}

            <TextInput
              required
              label="Email"
              placeholder="example@gmail.com"
              value={form.values.email}
              onChange={(event) =>
                form.setFieldValue('email', event.currentTarget.value)
              }
              error={form.errors.email && 'Invalid email'}
              radius="md"
            />

            <PasswordInput
              required
              label="Password"
              placeholder="Your password"
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

            {type === 'login' && (
              <Anchor
                component="a"
                href="/forgot-password"
                size="sm"
                c="dimmed"
                mt={5}
              >
                Forgot password?
              </Anchor>
            )}

            {type === 'register' && (
              <Checkbox
                label="I accept terms and conditions"
                color="green"
                checked={form.values.terms}
                onChange={(event) =>
                  form.setFieldValue('terms', event.currentTarget.checked)
                }
              />
            )}
          </Stack>

          <Group justify="space-between" mt="xl">
            <Anchor
              component="button"
              type="button"
              c="dimmed"
              onClick={() => toggle()}
              size="xs"
            >
              {type === 'register'
                ? 'Already have an account? Login'
                : "Don't have an account? Register"}
            </Anchor>
            <Button
              type="submit"
              bg="green"
              disabled={!form.values.terms}
              radius="xl"
              loading={loginMutation.isPending}
            >
              {upperFirst(type)}
            </Button>
          </Group>
        </form>
      </Paper>
    </div>
  )
}

export default Auth
