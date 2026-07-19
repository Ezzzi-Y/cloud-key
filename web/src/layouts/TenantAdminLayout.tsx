import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { LayoutDashboard, Key, Search, FileText, Server, User, LogOut, Menu } from 'lucide-react'
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet'
import { useState } from 'react'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/tenant', icon: LayoutDashboard, label: '仪表盘', end: true },
  { to: '/tenant/keys', icon: Key, label: 'Key 管理' },
  { to: '/tenant/keys/verify', icon: Search, label: '校验与扣减' },
  { to: '/tenant/logs', icon: FileText, label: '使用记录' },
  { to: '/tenant/service-accounts', icon: Server, label: '服务账号' },
  { to: '/tenant/profile', icon: User, label: '个人设置' },
]

export default function TenantAdminLayout() {
  const { username, logout } = useAuth()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const NavLinks = () => (
    <nav className="flex flex-col gap-1 px-2">
      {navItems.map(({ to, icon: Icon, label, end }) => (
        <NavLink
          key={to}
          to={to}
          end={end}
          onClick={() => setOpen(false)}
          className={({ isActive }) =>
            cn(
              'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
              isActive ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
            )
          }
        >
          <Icon className="h-4 w-4" />
          {label}
        </NavLink>
      ))}
    </nav>
  )

  return (
    <div className="flex h-screen">
      <aside className="hidden w-60 flex-col border-r bg-card md:flex">
        <div className="flex h-14 items-center border-b px-4">
          <h1 className="text-lg font-bold">CloudKey</h1>
          <span className="ml-2 rounded bg-blue-50 px-2 py-0.5 text-xs text-blue-600">租户管理员</span>
        </div>
        <div className="flex-1 overflow-auto py-4"><NavLinks /></div>
        <div className="border-t p-4">
          <div className="mb-2 text-sm text-muted-foreground">{username}</div>
          <Button variant="outline" size="sm" className="w-full" onClick={handleLogout}>
            <LogOut className="mr-2 h-4 w-4" />退出登录
          </Button>
        </div>
      </aside>

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild>
          <Button variant="ghost" size="icon" className="absolute left-4 top-3 md:hidden">
            <Menu className="h-5 w-5" />
          </Button>
        </SheetTrigger>
        <SheetContent side="left" className="w-60 p-0">
          <div className="flex h-14 items-center border-b px-4"><h1 className="text-lg font-bold">CloudKey</h1></div>
          <div className="py-4"><NavLinks /></div>
        </SheetContent>
      </Sheet>

      <main className="flex-1 overflow-auto">
        <div className="mx-auto max-w-7xl p-6"><Outlet /></div>
      </main>
    </div>
  )
}
