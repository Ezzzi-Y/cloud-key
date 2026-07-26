import AdminLayout from './AdminLayout'
import type { NavItem } from './AdminLayout'
import { useAuth } from '@/hooks/useAuth'
import TenantStatusBanner from '@/components/TenantStatusBanner'
import TenantBlockedPage from '@/components/TenantBlockedPage'
import { LayoutDashboard, Key, Search, FileText, ArrowLeftRight, Server, User } from 'lucide-react'

const navItems: NavItem[] = [
  { to: '/tenant', icon: LayoutDashboard, label: '仪表盘', end: true },
  { to: '/tenant/keys', icon: Key, label: 'Key 管理', end: true, writeOp: true },
  { to: '/tenant/keys/verify', icon: Search, label: '校验与扣减', writeOp: true },
  { to: '/tenant/logs', icon: FileText, label: '使用记录' },
  { to: '/tenant/balance-logs', icon: ArrowLeftRight, label: '额度调整记录' },
  { to: '/tenant/service-accounts', icon: Server, label: '服务账号', writeOp: true },
  { to: '/tenant/profile', icon: User, label: '个人设置' },
]

export default function TenantAdminLayout() {
  const { role, tenantStatus } = useAuth()

  // 等待 profile 请求完成，避免子页面提前发起 API 请求
  if (!role) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  if (tenantStatus === 'disabled') {
    return <TenantBlockedPage />
  }

  return (
    <AdminLayout
      navItems={navItems}
      roleLabel="租户管理员"
      roleColor="bg-blue-50 text-blue-600 dark:bg-blue-950 dark:text-blue-400"
      disableWriteOps={tenantStatus === 'expired'}
      alertSlot={<TenantStatusBanner />}
    />
  )
}
