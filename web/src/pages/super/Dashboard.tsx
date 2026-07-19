import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { listTenants } from '@/api/tenants'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import { SkeletonCards } from '@/components/SkeletonCard'
import { Building2, CheckCircle, Clock, XCircle, ExternalLink } from 'lucide-react'

interface Tenant {
  id: number
  name: string
  status: string
  expire_at: string
  key_count: number
  user_count: number
}

export default function SuperDashboard() {
  const { data, isLoading } = useQuery({
    queryKey: ['super-tenants'],
    queryFn: () => listTenants().then((r) => (r.code === 0 ? r.data : [])),
  })
  const tenants = (data as Tenant[]) || []
  const activeCount = tenants.filter((t) => t.status === 'active').length
  const expiredCount = tenants.filter((t) => t.status === 'expired').length
  const disabledCount = tenants.filter((t) => t.status === 'disabled').length

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-32" />
        <SkeletonCards />
        <Card><CardContent className="p-6"><Skeleton className="h-64 w-full" /></CardContent></Card>
      </div>
    )
  }

  const kpiCards = [
    { title: '总租户数', value: tenants.length, icon: Building2, color: 'text-blue-600 dark:text-blue-400' },
    { title: '活跃', value: activeCount, icon: CheckCircle, color: 'text-green-600 dark:text-green-400' },
    { title: '已过期', value: expiredCount, icon: Clock, color: 'text-yellow-600 dark:text-yellow-400' },
    { title: '已禁用', value: disabledCount, icon: XCircle, color: 'text-red-600 dark:text-red-400' },
  ]

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold tracking-tight">仪表盘</h2>

      <div className="grid gap-4 md:grid-cols-4">
        {kpiCards.map((card) => (
          <Card key={card.title} className="transition-shadow hover:shadow-md">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">{card.title}</CardTitle>
              <card.icon className={`h-5 w-5 ${card.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold tracking-tight">{card.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader><CardTitle className="text-base">租户列表</CardTitle></CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>到期时间</TableHead>
                <TableHead className="text-right">Key 数量</TableHead>
                <TableHead className="text-right">用户数</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tenants.map((t) => (
                <TableRow key={t.id} className="group">
                  <TableCell className="font-medium">{t.name}</TableCell>
                  <TableCell>
                    <Badge variant={t.status === 'active' ? 'success' : t.status === 'expired' ? 'warning' : 'destructive'}>
                      {t.status === 'active' ? '活跃' : t.status === 'expired' ? '已过期' : '已禁用'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">{t.expire_at || '永不过期'}</TableCell>
                  <TableCell className="text-right font-mono">{t.key_count}</TableCell>
                  <TableCell className="text-right font-mono">{t.user_count}</TableCell>
                  <TableCell className="text-right">
                    <Link
                      to={`/super/tenants/${t.id}`}
                      className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
                    >
                      详情 <ExternalLink className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
              {tenants.length === 0 && (
                <TableRow><TableCell colSpan={6} className="text-center text-muted-foreground py-8">暂无租户数据</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
