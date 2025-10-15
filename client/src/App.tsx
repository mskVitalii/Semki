import { Route, Routes } from 'react-router-dom'
import { NotFound } from './pages/404/NotFound'
import { Auth } from './pages/auth/Auth'
import { ForgotPassword } from './pages/auth/ForgotPassword'
import Landing from './pages/landing/Landing'
import Organization from './pages/organization/Organization'
import Profile from './pages/profile/Profile'
import QA from './pages/qa/QA'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/login" element={<Auth />} />
      <Route path="/profile/:userId" element={<Profile />} />
      <Route path="/organization" element={<Organization />} />
      <Route path="/qa" element={<QA />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}

export default App
