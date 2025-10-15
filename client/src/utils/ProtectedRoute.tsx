import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'

export const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const accessToken = useAuthStore((state) => state.accessToken)

  if (!accessToken) return <Navigate to="/login" />
  return <>{children}</>
}
