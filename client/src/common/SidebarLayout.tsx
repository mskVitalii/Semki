import { Box, Burger, Drawer } from '@mantine/core'
import { useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { Sidebar } from './Sidebar'

interface MainLayoutProps {
  children: ReactNode
}

export function MainLayout({ children }: MainLayoutProps) {
  // const userId = useAuthStore((state) => state.user?.id)
  const [opened, setOpened] = useState(false)
  const navigate = useNavigate()

  const handleNewChat = () => {
    console.log('new!')
    navigate('/qa', { replace: true })
  }

  // if (!userId) {
  //   return null
  // }

  return (
    <div className="flex h-screen overflow-hidden">
      <Box className="hidden! md:block! w-80">
        <Sidebar onNewChat={handleNewChat} />
      </Box>
      <Box className="flex-1 relative overflow-auto">
        <div className="md:hidden! fixed! top-4! left-4! z-50!">
          <Burger opened={opened} onClick={() => setOpened((o) => !o)} />
        </div>
        <Drawer
          opened={opened}
          onClose={() => setOpened(false)}
          size="80%"
          padding="md"
          withCloseButton={false}
          overlayProps={{ opacity: 0.4, blur: 2 }}
        >
          <Sidebar onNewChat={handleNewChat} />
        </Drawer>

        <main className="flex-1 flex items-center justify-center">
          {children}
        </main>
      </Box>
    </div>
  )
}
