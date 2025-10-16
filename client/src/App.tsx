import { lazy } from 'react'
import { Route, Routes } from 'react-router-dom'
import { BootstrapRoute } from './common/BootstrapRoute'
import { ProtectedRoute } from './common/ProtectedRoute'

const NotFound = lazy(() => import('./pages/404/NotFound'))
const Landing = lazy(() => import('./pages/landing/Landing'))
const Organization = lazy(() => import('./pages/organization/Organization'))
const Profile = lazy(() => import('./pages/profile/Profile'))
const QA = lazy(() => import('./pages/qa/QA'))
const Auth = lazy(() => import('./pages/auth/Auth'))
const ForgotPassword = lazy(() => import('./pages/auth/ForgotPassword'))

function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route element={<BootstrapRoute />}>
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/login" element={<Auth />} />
        <Route element={<ProtectedRoute />}>
          <Route path="/profile/:userId" element={<Profile />} />
          <Route path="/organization" element={<Organization />} />
          <Route path="/qa" element={<QA />} />
          <Route path="/qa/:chatId" element={<QA />} />
        </Route>
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}

export default App
