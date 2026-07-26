import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import type { UserRole, TenantStatus } from '@/types'
import { getProfile } from '@/api/auth'

interface AuthState {
  token: string | null
  role: UserRole | null
  tenantId: number | null
  username: string | null
  isAuthenticated: boolean
  tenantStatus: TenantStatus | null
  tenantExpireAt: string | null
}

interface AuthContextType extends AuthState {
  login: (token: string, role: UserRole, tenantId: number | null, username: string) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(() => {
    const token = localStorage.getItem('ck_token')
    return { token, role: null, tenantId: null, username: null, isAuthenticated: !!token, tenantStatus: null, tenantExpireAt: null }
  })
  const navigate = useNavigate()

  useEffect(() => {
    const token = localStorage.getItem('ck_token')
    if (token && !state.role) {
      getProfile('super')
        .then((res) => {
          if (res.code === 0) {
            const d = res.data
            setState({ token, role: d.role, tenantId: d.tenant_id ?? null, username: d.username, isAuthenticated: true, tenantStatus: null, tenantExpireAt: null })
          }
        })
        .catch(() => {
          getProfile('tenant')
            .then((res2) => {
              if (res2.code === 0) {
                const d = res2.data
                setState({
                  token, role: d.role, tenantId: d.tenant_id ?? null, username: d.username, isAuthenticated: true,
                  tenantStatus: d.tenant_status ?? null,
                  tenantExpireAt: d.tenant_expire_at ?? null,
                })
              }
            })
            .catch((err) => {
              const code = err?.code
              if (code === 4001) {
                // expired：保持登录态，前端可展示 Banner 和只读数据
                setState({
                  token, role: 'tenant_admin', tenantId: null, username: null, isAuthenticated: true,
                  tenantStatus: 'expired', tenantExpireAt: null,
                })
                return
              }
              if (code === 4002) {
                // disabled：保持登录态，由布局层展示全屏拦截页
                setState({
                  token, role: 'tenant_admin', tenantId: null, username: null, isAuthenticated: true,
                  tenantStatus: 'disabled', tenantExpireAt: null,
                })
                return
              }
              // 其他错误（网络/服务器错误）：清除认证态，跳转登录
              localStorage.removeItem('ck_token')
              setState({ token: null, role: null, tenantId: null, username: null, isAuthenticated: false, tenantStatus: null, tenantExpireAt: null })
              navigate('/login')
            })
        })
    }
  }, [])

  const login = useCallback((token: string, role: UserRole, tenantId: number | null, username: string) => {
    localStorage.setItem('ck_token', token)
    setState({ token, role, tenantId, username, isAuthenticated: true, tenantStatus: null, tenantExpireAt: null })
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('ck_token')
    setState({ token: null, role: null, tenantId: null, username: null, isAuthenticated: false, tenantStatus: null, tenantExpireAt: null })
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
