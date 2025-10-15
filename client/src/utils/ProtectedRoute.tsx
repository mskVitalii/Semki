import { useAuthStore } from '@/stores/authStore'
import { Navigate, Outlet } from 'react-router-dom'

export const ProtectedRoute = ({
  children,
}: {
  children?: React.ReactNode
}) => {
  const accessToken = useAuthStore((state) => state.accessToken)

  if (!accessToken) return <Navigate to="/login" replace />

  return children ? <>{children}</> : <Outlet />
}
