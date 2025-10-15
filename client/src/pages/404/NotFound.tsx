import { Anchor, Button, Container, Group, Text, Title } from '@mantine/core'
import { Illustration } from './Illustration'

export function NotFound() {
  return (
    <div className="flex items-center justify-center min-h-screen max-w-screen w-screen bg-gray-800">
      <div className="relative w-full h-screen overflow-hidden bg-gray-900 text-white">
        <div className="absolute inset-0 whitespace-nowrap animate-scroll text-[10rem] font-extrabold opacity-10">
          {'404 '.repeat(20)}
        </div>
        <div className="absolute bottom-0 whitespace-nowrap animate-scroll text-[10rem] font-extrabold opacity-10">
          {'404 '.repeat(20)}
        </div>
        <div className="absolute inset-0 rotate-90 whitespace-nowrap animate-scroll text-[10rem] font-extrabold opacity-10">
          {'404 '.repeat(20)}
        </div>
        <div className="absolute inset-0 rotate-270 whitespace-nowrap animate-scroll text-[10rem] font-extrabold opacity-10">
          {'404 '.repeat(20)}
        </div>

        <div className="relative z-10 flex items-center justify-center h-full">
          <Container className="pt-20 pb-20 relative max-w-[450px]">
            <div className="relative">
              <Illustration className="absolute inset-0 opacity-75 text-gray-100 dark:text-gray-700" />
              <div className="relative z-10 pt-[220px] sm:pt-[120px] text-center flex flex-col items-center gap-[10vh]">
                <Title className="max-w-[450px] font-outfit font-medium text-[38px] sm:text-[32px] text-gray-100">
                  {getRandomExcuse()}
                </Title>
                <Text
                  size="lg"
                  ta="center"
                  className="max-w-[450px] mx-auto mt-8 mb-12 text-gray-200!"
                >
                  Page you are trying to open does not exist. You may have
                  mistyped the address, or the page has been moved to another
                  URL. If you think this is an error contact support
                  <Text component="span" display="block" ta="center">
                    <Anchor
                      href="mailto:msk.vitaly@gmail.com"
                      target="_blank"
                      underline="always"
                    >
                      msk.vitaly@gmail.com
                    </Anchor>
                  </Text>
                </Text>
                <Group justify="center">
                  <Button size="md" component="a" href="/">
                    Country roads, take me home
                  </Button>
                </Group>
              </div>
            </div>
          </Container>
        </div>
      </div>
    </div>
  )
}

//#region Excuses
// 404 page not found messages

const excuses = [
  'This page has wandered off into the digital wilderness 🌲💻',
  'Our server elves are on a coffee break ☕🧝',
  "This page took a detour to the internet's Bermuda Triangle 🛳️🌊",
  "The page you're seeking is currently out of office 🏖️📄",
  'This page has been abducted by aliens 👽🛸',
  'This page is on a secret mission 🕵️‍♂️✉️',
  "The page you're trying to reach is lost in cyberspace 🌌💾",
  'Our website gnomes are still building this page 🏗️🧙‍♂️',
  'This page has gone fishing 🎣🌊',
]

function getRandomExcuse() {
  const index = Math.floor(Math.random() * excuses.length)
  return excuses[index]
}
//#endregion

export default NotFound
