import { Routes, Route, Navigate } from 'react-router-dom'
import { Suspense, lazy } from 'react'
import { AuthProvider } from '@/hooks/useAuth'
import RequireAuth from '@/components/RequireAuth'
import SuperAdminLayout from '@/layouts/SuperAdminLayout'
import TenantAdminLayout from '@/layouts/TenantAdminLayout'
import { Toaster } from '@/components/ui/sonner'

const Login = lazy(() => import('@/pages/Login'))
const SuperDashboard = lazy(() => import('@/pages/super/Dashboard'))
const Tenants = lazy(() => import('@/pages/super/Tenants'))
const TenantDetail = lazy(() => import('@/pages/super/TenantDetail'))
const PlatformConfig = lazy(() => import('@/pages/super/PlatformConfig'))
const SuperProfile = lazy(() => import('@/pages/super/Profile'))
const TenantDashboard = lazy(() => import('@/pages/tenant/Dashboard'))
const KeyManagement = lazy(() => import('@/pages/tenant/KeyManagement'))
const KeyVerify = lazy(() => import('@/pages/tenant/KeyVerify'))
const UsageLogs = lazy(() => import('@/pages/tenant/UsageLogs'))
const ServiceAccounts = lazy(() => import('@/pages/tenant/ServiceAccounts'))
const TenantProfile = lazy(() => import('@/pages/tenant/Profile'))
const NotFound = lazy(() => import('@/pages/NotFound'))

const Loading = () => (
  <div className="flex h-screen items-center justify-center">
    <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
  </div>
)

export default function App() {
  return (
    <AuthProvider>
      <Suspense fallback={<Loading />}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/super" element={<RequireAuth role="super_admin" />}>
            <Route element={<SuperAdminLayout />}>
              <Route index element={<SuperDashboard />} />
              <Route path="tenants" element={<Tenants />} />
              <Route path="tenants/:id" element={<TenantDetail />} />
              <Route path="config" element={<PlatformConfig />} />
              <Route path="profile" element={<SuperProfile />} />
            </Route>
          </Route>
          <Route path="/tenant" element={<RequireAuth role="tenant_admin" />}>
            <Route element={<TenantAdminLayout />}>
              <Route index element={<TenantDashboard />} />
              <Route path="keys" element={<KeyManagement />} />
              <Route path="keys/verify" element={<KeyVerify />} />
              <Route path="logs" element={<UsageLogs />} />
              <Route path="service-accounts" element={<ServiceAccounts />} />
              <Route path="profile" element={<TenantProfile />} />
            </Route>
          </Route>
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Suspense>
      <Toaster position="top-right" richColors />
    </AuthProvider>
  )
}
