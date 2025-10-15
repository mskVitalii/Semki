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
import { IconArrowLeft } from '@tabler/icons-react'

// TODO: request
// TODO: notification

export function ForgotPassword() {
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
          <TextInput
            label="Your email"
            defaultValue={import.meta.env.VITE_TEST_EMAIL ?? ''}
            placeholder="example@gmail.com"
            required
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
              bg="green"
              className="bg-blue-500 text-white px-4 py-2 rounded w-full sm:w-auto"
            >
              Reset password
            </Button>
          </Group>
        </Paper>
      </Container>
    </div>
  )
}

export default ForgotPassword
