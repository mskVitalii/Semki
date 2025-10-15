import {
  Anchor,
  Button,
  Container,
  Group,
  Image,
  Text,
  Title,
} from '@mantine/core'
import image from '../media/image_hero.svg'
import Logo from './Logo'

export function Hero() {
  return (
    <Container size="md" className="min-h-screen">
      <div className="flex min-h-screen justify-between items-center pt-[calc(var(--mantine-spacing-xl)*4)] pb-[calc(var(--mantine-spacing-xl)*4)]">
        <div className="max-w-[480px] mr-[calc(var(--mantine-spacing-xl)*3)] md:max-w-full md:mr-0">
          <Logo />
          <Title className="text-[light-dark(var(--mantine-color-black),var(--mantine-color-white))] font-[Outfit,var(--mantine-font-family)] text-[44px] leading-[1.2] font-medium xs:text-[28px]">
            A
            <span className="relative bg-[var(--mantine-color-green-light)] rounded-[var(--mantine-radius-sm)] py-2 px-3">
              {' '}
              semantic{' '}
            </span>
            Tool <br /> for employee search
          </Title>
          <Text c="dimmed" mt="md">
            Connect people across departments in seconds
          </Text>

          <Group mt={30}>
            <Anchor href="/login">
              <Button radius="xl" bg="green" size="md" className="xs:flex-1">
                Get started
              </Button>
            </Anchor>
          </Group>
        </div>
        <Image src={image} className="w-[50%]! aspect-square md:hidden" />
      </div>
    </Container>
  )
}
