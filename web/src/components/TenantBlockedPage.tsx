import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { ShieldOff, LogOut } from 'lucide-react'

export default function TenantBlockedPage() {
  const { logout } = useAuth()

  return (
    <div className="flex h-screen items-center justify-center bg-muted/30 px-4">
      <div className="flex flex-col items-center space-y-6 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
          <ShieldOff className="h-8 w-8 text-destructive" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold text-foreground">租户已被禁用</h1>
          <p className="max-w-md text-sm text-muted-foreground">
            当前租户已被管理员禁用，所有操作已暂停。如需恢复访问，请联系平台管理员。
          </p>
        </div>
        <Button variant="outline" onClick={logout}>
          <LogOut className="mr-2 h-4 w-4" />
          退出登录
        </Button>
      </div>
    </div>
  )
}
