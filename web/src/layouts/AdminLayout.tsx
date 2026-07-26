import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet'
import { ThemeToggle } from '@/components/ThemeToggle'
import { LogOut, Menu, PanelLeftClose, PanelLeft, type LucideIcon } from 'lucide-react'
import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'

export interface NavItem {
  to: string
  icon: LucideIcon
  label: string
  end?: boolean
  writeOp?: boolean
}

interface AdminLayoutProps {
  navItems: NavItem[]
  roleLabel: string
  roleColor: string
  disableWriteOps?: boolean
  alertSlot?: React.ReactNode
}

export default function AdminLayout({ navItems, roleLabel, roleColor, disableWriteOps, alertSlot }: AdminLayoutProps) {
  const { username, logout } = useAuth()
  const navigate = useNavigate()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [collapsed, setCollapsed] = useState(() => localStorage.getItem('sidebar-collapsed') === 'true')

  useEffect(() => {
    localStorage.setItem('sidebar-collapsed', String(collapsed))
  }, [collapsed])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userInitial = username ? username.charAt(0).toUpperCase() : 'U'

  const NavLinks = ({ compact = false }: { compact?: boolean }) => (
    <nav className="flex flex-col gap-1 px-2">
      {navItems.map(({ to, icon: Icon, label, end, writeOp }) => {
        const disabled = disableWriteOps && writeOp
        const link = (
          <NavLink
            key={to}
            to={to}
            end={end}
            onClick={(e) => { if (disabled) { e.preventDefault(); return } setMobileOpen(false) }}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                compact && 'justify-center px-0',
                disabled && 'pointer-events-none opacity-40',
                isActive && !disabled
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
              )
            }
          >
            <Icon className="h-4 w-4 shrink-0" />
            {!compact && <span className="truncate">{label}</span>}
          </NavLink>
        )

        if (compact) {
          return (
            <Tooltip key={to} delayDuration={0}>
              <TooltipTrigger asChild>{link}</TooltipTrigger>
              <TooltipContent side="right">{disabled ? '当前租户状态不可操作' : label}</TooltipContent>
            </Tooltip>
          )
        }
        return link
      })}
    </nav>
  )

  const sidebarWidth = collapsed ? 'w-16' : 'w-60'

  return (
    <TooltipProvider>
      <div className="flex h-screen">
        {/* Desktop sidebar */}
        <aside className={cn('hidden flex-col border-r bg-card transition-[width] duration-200 md:flex', sidebarWidth)}>
          <div className="flex h-14 items-center border-b px-4">
            {!collapsed && (
              <div className="flex items-center gap-2">
                <h1 className="text-lg font-bold">CloudKey</h1>
                <span className={cn('rounded px-1.5 py-0.5 text-[10px] font-medium', roleColor)}>
                  {roleLabel}
                </span>
              </div>
            )}
            {collapsed && <h1 className="text-lg font-bold mx-auto">CK</h1>}
          </div>

          <ScrollArea className="flex-1 py-4">
            <NavLinks compact={collapsed} />
          </ScrollArea>

          <Separator />

          <div className="p-3">
            {!collapsed ? (
              <div className="flex items-center gap-3">
                <Avatar className="h-8 w-8">
                  <AvatarFallback className="bg-primary/10 text-primary text-xs font-medium">
                    {userInitial}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1 min-w-0">
                  <div className="truncate text-sm font-medium">{username}</div>
                </div>
                <ThemeToggle />
              </div>
            ) : (
              <div className="flex flex-col items-center gap-2">
                <Avatar className="h-8 w-8">
                  <AvatarFallback className="bg-primary/10 text-primary text-xs font-medium">
                    {userInitial}
                  </AvatarFallback>
                </Avatar>
                <ThemeToggle />
              </div>
            )}
          </div>

          <Separator />

          <div className="p-3">
            {!collapsed ? (
              <Button variant="outline" size="sm" className="w-full" onClick={handleLogout}>
                <LogOut className="mr-2 h-4 w-4" />
                退出登录
              </Button>
            ) : (
              <Tooltip delayDuration={0}>
                <TooltipTrigger asChild>
                  <Button variant="outline" size="icon" className="w-full h-8" onClick={handleLogout}>
                    <LogOut className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="right">退出登录</TooltipContent>
              </Tooltip>
            )}
          </div>
        </aside>

        {/* Mobile hamburger */}
        <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="absolute left-4 top-3 md:hidden z-50">
              <Menu className="h-5 w-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-60 p-0">
            <div className="flex h-14 items-center border-b px-4">
              <h1 className="text-lg font-bold">CloudKey</h1>
              <span className={cn('ml-2 rounded px-1.5 py-0.5 text-[10px] font-medium', roleColor)}>
                {roleLabel}
              </span>
            </div>
            <ScrollArea className="py-4 h-[calc(100vh-3.5rem)]">
              <NavLinks />
            </ScrollArea>
          </SheetContent>
        </Sheet>

        {/* Main content */}
        <main className="flex-1 overflow-auto">
          {/* Top bar with collapse toggle */}
          <div className="sticky top-0 z-10 flex h-14 items-center border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-6">
            <Button
              variant="ghost"
              size="icon"
              className="hidden md:flex h-8 w-8 mr-2"
              onClick={() => setCollapsed(!collapsed)}
            >
              {collapsed ? <PanelLeft className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
            </Button>
            <div className="flex-1" />
            <div className="flex items-center gap-2 md:hidden">
              <ThemeToggle />
            </div>
          </div>
          <div className="mx-auto max-w-7xl p-6">
            {alertSlot}
            <Outlet />
          </div>
        </main>
      </div>
    </TooltipProvider>
  )
}
