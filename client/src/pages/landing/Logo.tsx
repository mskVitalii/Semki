import { Anchor, Group, Title } from '@mantine/core'

function Logo() {
  return (
    <Group align="center" mb="lg">
      <Anchor
        variant="gradient"
        gradient={{ from: 'green', to: 'white' }}
        fw={500}
        fz="h1"
        href="/"
      >
        <Title order={1}>🌻 Semki</Title>
      </Anchor>
    </Group>
  )
}

export default Logo
