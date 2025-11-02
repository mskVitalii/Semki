import { lazy } from 'react'
import { Route, Routes } from 'react-router-dom'
import { BootstrapRoute } from './common/BootstrapRoute'
import { ProtectedRoute } from './common/ProtectedRoute'
import Onboarding from './pages/onboarding/Onboarding'

const NotFound = lazy(() => import('./pages/404/NotFound'))
const Landing = lazy(() => import('./pages/landing/Landing'))
const Organization = lazy(() => import('./pages/organization/Organization'))
const Profile = lazy(() => import('./pages/profile/Profile'))
const Chat = lazy(() => import('./pages/chat/Chat'))
const Auth = lazy(() => import('./pages/auth/Auth'))
const ForgotPassword = lazy(() => import('./pages/auth/ForgotPassword'))

function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/login" element={<Auth />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/onboarding" element={<Onboarding />} />
      <Route element={<BootstrapRoute />}>
        <Route element={<ProtectedRoute />}>
          <Route path="/profile/:userId" element={<Profile />} />
          <Route path="/organization" element={<Organization />} />
          <Route path="/chat" element={<Chat />} />
          <Route path="/chat/:chatId" element={<Chat />} />
        </Route>
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}

export default App
