import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import type { UserRole } from '@/types'
import { getProfile } from '@/api/auth'

interface AuthState {
  token: string | null
  role: UserRole | null
  tenantId: number | null
  username: string | null
  isAuthenticated: boolean
}

interface AuthContextType extends AuthState {
  login: (token: string, role: UserRole, tenantId: number | null, username: string) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(() => {
    const token = localStorage.getItem('ck_token')
    return { token, role: null, tenantId: null, username: null, isAuthenticated: !!token }
  })
  const navigate = useNavigate()

  useEffect(() => {
    const token = localStorage.getItem('ck_token')
    if (token && !state.role) {
      getProfile('super')
        .then((res) => {
          if (res.code === 0) {
            const d = res.data
            setState({ token, role: d.role, tenantId: d.tenant_id ?? null, username: d.username, isAuthenticated: true })
          }
        })
        .catch(() => {
          getProfile('tenant')
            .then((res2) => {
              if (res2.code === 0) {
                const d = res2.data
                setState({ token, role: d.role, tenantId: d.tenant_id ?? null, username: d.username, isAuthenticated: true })
              }
            })
            .catch(() => {
              localStorage.removeItem('ck_token')
              setState({ token: null, role: null, tenantId: null, username: null, isAuthenticated: false })
              navigate('/login')
            })
        })
    }
  }, [])

  const login = useCallback((token: string, role: UserRole, tenantId: number | null, username: string) => {
    localStorage.setItem('ck_token', token)
    setState({ token, role, tenantId, username, isAuthenticated: true })
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('ck_token')
    setState({ token: null, role: null, tenantId: null, username: null, isAuthenticated: false })
    navigate('/login')
  }, [navigate])

  return (
    <AuthContext.Provider value={{ ...state, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
