import { useAuth } from '@/hooks/useAuth'
import { AlertTriangle } from 'lucide-react'

export default function TenantStatusBanner() {
  const { tenantStatus, tenantExpireAt } = useAuth()

  if (tenantStatus === 'expired') {
    return (
      <div className="mb-6 flex items-start gap-3 rounded-lg border border-yellow-200 bg-yellow-50 p-4 dark:border-yellow-800 dark:bg-yellow-950">
        <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-yellow-600 dark:text-yellow-400" />
        <p className="text-sm text-yellow-800 dark:text-yellow-200">
          租户已到期，仅可查看统计数据。如需续期请联系管理员。
        </p>
      </div>
    )
  }

  // 即将到期预警（7 天内）
  if (tenantStatus === 'active' && tenantExpireAt) {
    const expireDate = new Date(tenantExpireAt)
    const now = new Date()
    const daysLeft = Math.ceil((expireDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    if (daysLeft > 0 && daysLeft <= 7) {
      return (
        <div className="mb-6 flex items-start gap-3 rounded-lg border border-yellow-200 bg-yellow-50 p-4 dark:border-yellow-800 dark:bg-yellow-950">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-yellow-600 dark:text-yellow-400" />
          <p className="text-sm text-yellow-800 dark:text-yellow-200">
            租户将于 {daysLeft} 天后到期，请及时续期以免影响使用。
          </p>
        </div>
      )
    }
  }

  return null
}
