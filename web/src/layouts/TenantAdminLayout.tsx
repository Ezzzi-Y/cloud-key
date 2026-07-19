import AdminLayout from './AdminLayout'
import type { NavItem } from './AdminLayout'
import { LayoutDashboard, Key, Search, FileText, Server, User } from 'lucide-react'

const navItems: NavItem[] = [
  { to: '/tenant', icon: LayoutDashboard, label: '仪表盘', end: true },
  { to: '/tenant/keys', icon: Key, label: 'Key 管理', end: true },
  { to: '/tenant/keys/verify', icon: Search, label: '校验与扣减' },
  { to: '/tenant/logs', icon: FileText, label: '使用记录' },
  { to: '/tenant/service-accounts', icon: Server, label: '服务账号' },
  { to: '/tenant/profile', icon: User, label: '个人设置' },
]

export default function TenantAdminLayout() {
  return (
    <AdminLayout
      navItems={navItems}
      roleLabel="租户管理员"
      roleColor="bg-blue-50 text-blue-600 dark:bg-blue-950 dark:text-blue-400"
    />
  )
}
