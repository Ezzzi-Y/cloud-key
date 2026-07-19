import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { listTenants } from '@/api/tenants'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Building2, CheckCircle, Clock, XCircle } from 'lucide-react'

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

  if (isLoading) return <div className="flex items-center justify-center py-12">加载中...</div>

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">仪表盘</h2>
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">总租户数</CardTitle>
            <Building2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{tenants.length}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">活跃</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold text-green-600">{activeCount}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">已过期</CardTitle>
            <Clock className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold text-yellow-600">{expiredCount}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">已禁用</CardTitle>
            <XCircle className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold text-red-600">{disabledCount}</div></CardContent>
        </Card>
      </div>
      <Card>
        <CardHeader><CardTitle>租户列表</CardTitle></CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>到期时间</TableHead>
                <TableHead>Key 数量</TableHead>
                <TableHead>用户数</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tenants.map((t) => (
                <TableRow key={t.id}>
                  <TableCell className="font-medium">{t.name}</TableCell>
                  <TableCell>
                    <Badge variant={t.status === 'active' ? 'success' : t.status === 'expired' ? 'warning' : 'destructive'}>
                      {t.status === 'active' ? '活跃' : t.status === 'expired' ? '已过期' : '已禁用'}
                    </Badge>
                  </TableCell>
                  <TableCell>{t.expire_at || '永不过期'}</TableCell>
                  <TableCell>{t.key_count}</TableCell>
                  <TableCell>{t.user_count}</TableCell>
                  <TableCell>
                    <Link to={`/super/tenants/${t.id}`} className="text-primary hover:underline">详情</Link>
                  </TableCell>
                </TableRow>
              ))}
              {tenants.length === 0 && (
                <TableRow><TableCell colSpan={6} className="text-center text-muted-foreground">暂无租户数据</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
