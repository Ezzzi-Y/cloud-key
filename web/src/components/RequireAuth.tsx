import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import type { UserRole } from '@/types'

interface Props {
  role: UserRole
}

export default function RequireAuth({ role }: Props) {
  const { isAuthenticated, role: currentRole } = useAuth()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  if (currentRole !== role) {
    const correctPath = currentRole === 'super_admin' ? '/super/' : '/tenant/'
    return <Navigate to={correctPath} replace />
  }

  return <Outlet />
}
