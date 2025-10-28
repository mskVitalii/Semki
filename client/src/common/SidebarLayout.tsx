import { Box, Burger, Drawer } from '@mantine/core'
import { useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { Sidebar } from './Sidebar'

interface MainLayoutProps {
  children: ReactNode
}

export function MainLayout({ children }: MainLayoutProps) {
  const [opened, setOpened] = useState(false)
  const navigate = useNavigate()

  const handleNewChat = () => {
    navigate('/chat', { replace: true })
  }

  return (
    <div className="flex h-screen overflow-x-hidden">
      <Box className="hidden! md:block! w-80">
        <Sidebar onNewChat={handleNewChat} />
      </Box>
      <Box className="flex-1 relative">
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
          classNames={{ root: 'md:hidden' }}
        >
          <Sidebar onNewChat={handleNewChat} />
        </Drawer>

        <main className="flex-1 flex items-start justify-center h-full max-w-screen overflow-auto">
          {children}
        </main>
      </Box>
    </div>
  )
}
