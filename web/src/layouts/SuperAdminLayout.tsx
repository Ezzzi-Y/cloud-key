import AdminLayout from './AdminLayout'
import type { NavItem } from './AdminLayout'
import { LayoutDashboard, Building2, Settings, User } from 'lucide-react'

const navItems: NavItem[] = [
  { to: '/super', icon: LayoutDashboard, label: '仪表盘', end: true },
  { to: '/super/tenants', icon: Building2, label: '租户管理' },
  { to: '/super/config', icon: Settings, label: '平台配置' },
  { to: '/super/profile', icon: User, label: '个人设置' },
]

export default function SuperAdminLayout() {
  return (
    <AdminLayout
      navItems={navItems}
      roleLabel="超级管理员"
      roleColor="bg-primary/10 text-primary"
    />
  )
}
