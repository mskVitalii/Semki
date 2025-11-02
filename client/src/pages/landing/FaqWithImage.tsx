import image from '@/media/image.svg'
import { Accordion, Anchor, Container, Grid, Image, Title } from '@mantine/core'

const FAQ = [
  {
    question: 'Would you like to hire me? :)',
    answer: (
      <Anchor
        href="mailto:msk.vitaly@gmail.com"
        target="_blank"
        underline="always"
      >
        msk.vitaly@gmail.com
      </Anchor>
    ),
  },
]

export function FaqWithImage() {
  return (
    <div className="pt-[calc(var(--mantine-spacing-xl)*2)] pb-[calc(var(--mantine-spacing-xl)*2)]">
      <Container size="lg">
        <Grid id="faq-grid" gutter={50}>
          <Grid.Col
            span={{ base: 12, md: 6 }}
            className="flex justify-center items-center"
          >
            <Image src={image} alt="Frequently Asked Questions" />
          </Grid.Col>
          <Grid.Col span={{ base: 12, md: 6 }}>
            <Title
              order={2}
              ta="left"
              className="pl-[var(--mantine-spacing-md)] text-[var(--mantine-color-white)] font-[Outfit,var(--mantine-font-family)] font-medium mb-4!"
            >
              Frequently Asked Questions
            </Title>

            <Accordion
              chevronPosition="right"
              defaultValue="reset_password"
              variant="separated"
            >
              {FAQ.map(({ question, answer }) => (
                <Accordion.Item
                  key={question}
                  className="text-[var(--mantine-font-size-sm)]"
                  value="reset_password"
                >
                  <Accordion.Control>{question}</Accordion.Control>
                  <Accordion.Panel>{answer}</Accordion.Panel>
                </Accordion.Item>
              ))}
            </Accordion>
          </Grid.Col>
        </Grid>
      </Container>
    </div>
  )
}
