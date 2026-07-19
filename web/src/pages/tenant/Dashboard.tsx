import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getDashboard, getTrends, getTopKeys, getTopIPs } from '@/api/stats'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { Key, Calendar, TrendingUp, BarChart3 } from 'lucide-react'

export default function TenantDashboard() {
  const [period, setPeriod] = useState<'today' | 'week' | 'month'>('today')

  const { data: dash, isLoading } = useQuery({
    queryKey: ['tenant-dashboard'],
    queryFn: () => getDashboard().then((r) => (r.code === 0 ? r.data : null)),
  })
  const { data: trends } = useQuery({
    queryKey: ['tenant-trends', period],
    queryFn: () => getTrends(period).then((r) => (r.code === 0 ? r.data : null)),
  })
  const { data: topKeys } = useQuery({
    queryKey: ['tenant-top-keys'],
    queryFn: () => getTopKeys().then((r) => (r.code === 0 ? r.data : [])),
  })
  const { data: topIPs } = useQuery({
    queryKey: ['tenant-top-ips'],
    queryFn: () => getTopIPs().then((r) => (r.code === 0 ? r.data : [])),
  })

  if (isLoading) return <div className="flex items-center justify-center py-12">加载中...</div>

  const statusBreakdown = dash?.key_status_breakdown || {}

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">仪表盘</h2>
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Key 总数</CardTitle>
            <Key className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{dash?.key_count ?? 0}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">今日调用</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{dash?.today_calls ?? 0}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">本周调用</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{dash?.week_calls ?? 0}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">本月调用</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent><div className="text-2xl font-bold">{dash?.month_calls ?? 0}</div></CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader><CardTitle>Key 状态分布</CardTitle></CardHeader>
        <CardContent>
          <div className="grid grid-cols-4 gap-4 text-center">
            {Object.entries(statusBreakdown).map(([status, count]) => (
              <div key={status}>
                <div className="text-2xl font-bold">{count as number}</div>
                <div className="text-sm text-muted-foreground">
                  {status === 'unused' ? '未使用' : status === 'used' ? '已用尽' : status === 'disabled' ? '已禁用' : '已过期'}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>调用趋势</CardTitle>
            <div className="flex gap-2">
              {(['today', 'week', 'month'] as const).map((p) => (
                <Button key={p} variant={period === p ? 'default' : 'outline'} size="sm" onClick={() => setPeriod(p)}>
                  {p === 'today' ? '今日' : p === 'week' ? '本周' : '本月'}
                </Button>
              ))}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {trends?.points && trends.points.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={trends.points}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" fontSize={12} />
                <YAxis fontSize={12} />
                <Tooltip />
                <Bar dataKey="calls" fill="hsl(var(--primary))" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="py-8 text-center text-muted-foreground">暂无趋势数据</div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>Top 10 卡密</CardTitle></CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow><TableHead>排名</TableHead><TableHead>别名</TableHead><TableHead className="text-right">调用数</TableHead></TableRow>
              </TableHeader>
              <TableBody>
                {topKeys?.slice(0, 10).map((item, i) => (
                  <TableRow key={i}>
                    <TableCell>{i + 1}</TableCell>
                    <TableCell>{item.key_alias || item.key_suffix || '-'}</TableCell>
                    <TableCell className="text-right">{item.count}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>Top 10 IP</CardTitle></CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow><TableHead>排名</TableHead><TableHead>IP</TableHead><TableHead className="text-right">调用数</TableHead></TableRow>
              </TableHeader>
              <TableBody>
                {topIPs?.slice(0, 10).map((item, i) => (
                  <TableRow key={i}>
                    <TableCell>{i + 1}</TableCell>
                    <TableCell>{item.ip || '-'}</TableCell>
                    <TableCell className="text-right">{item.count}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader><CardTitle>最近使用记录</CardTitle></CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow><TableHead>Key 别名</TableHead><TableHead>扣减数量</TableHead><TableHead>IP</TableHead><TableHead>时间</TableHead></TableRow>
            </TableHeader>
            <TableBody>
              {dash?.recent_logs?.slice(0, 20).map((log) => (
                <TableRow key={log.id}>
                  <TableCell>{log.key_alias}</TableCell>
                  <TableCell>{log.amount}</TableCell>
                  <TableCell>{log.ip}</TableCell>
                  <TableCell>{new Date(log.created_at).toLocaleString('zh-CN')}</TableCell>
                </TableRow>
              ))}
              {(!dash?.recent_logs || dash.recent_logs.length === 0) && (
                <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground">暂无使用记录</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
