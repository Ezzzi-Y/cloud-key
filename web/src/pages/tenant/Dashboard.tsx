import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getDashboard, getTrends, getTopKeys } from '@/api/stats'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { SkeletonCards } from '@/components/SkeletonCard'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { Key, Calendar, TrendingUp, BarChart3, Plus } from 'lucide-react'
import { Link } from 'react-router-dom'

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

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-32" />
        <SkeletonCards />
        <Card><CardContent className="p-6"><Skeleton className="h-48 w-full" /></CardContent></Card>
        <Card><CardContent className="p-6"><Skeleton className="h-72 w-full" /></CardContent></Card>
      </div>
    )
  }

  const statusBreakdown = dash?.key_status_breakdown || {}
  const statusLabels: Record<string, string> = { active: '可用', exhausted: '已用尽', disabled: '已禁用', expired: '已过期' }
  const statusColors: Record<string, string> = {
    active: 'bg-green-100 text-green-700 dark:bg-green-950 dark:text-green-400',
    exhausted: 'bg-orange-100 text-orange-700 dark:bg-orange-950 dark:text-orange-400',
    disabled: 'bg-red-100 text-red-700 dark:bg-red-950 dark:text-red-400',
    expired: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-950 dark:text-yellow-400',
  }

  const kpiCards = [
    { title: 'Key 总数', value: dash?.key_count ?? 0, icon: Key, color: 'text-blue-600 dark:text-blue-400' },
    { title: '今日调用', value: dash?.today_calls ?? 0, icon: Calendar, color: 'text-green-600 dark:text-green-400' },
    { title: '本周调用', value: dash?.week_calls ?? 0, icon: TrendingUp, color: 'text-purple-600 dark:text-purple-400' },
    { title: '本月调用', value: dash?.month_calls ?? 0, icon: BarChart3, color: 'text-orange-600 dark:text-orange-400' },
  ]

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">仪表盘</h2>
        <Button asChild size="sm">
          <Link to="/tenant/keys"><Plus className="mr-2 h-4 w-4" />创建 Key</Link>
        </Button>
      </div>

      {/* KPI Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        {kpiCards.map((card) => (
          <Card key={card.title} className="transition-shadow hover:shadow-md">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">{card.title}</CardTitle>
              <card.icon className={`h-5 w-5 ${card.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold tracking-tight">{card.value.toLocaleString()}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Key Status Breakdown */}
      <Card>
        <CardHeader><CardTitle className="text-base">Key 状态分布</CardTitle></CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            {Object.entries(statusBreakdown).map(([status, count]) => (
              <div key={status} className="flex items-center gap-2">
                <Badge className={statusColors[status] || ''} variant="secondary">
                  {statusLabels[status] || status}
                </Badge>
                <span className="text-2xl font-bold">{count as number}</span>
              </div>
            ))}
            {Object.keys(statusBreakdown).length === 0 && (
              <p className="text-sm text-muted-foreground">暂无 Key 数据</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Trend Chart */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">调用趋势</CardTitle>
            <div className="flex gap-1 rounded-lg bg-muted p-1">
              {(['today', 'week', 'month'] as const).map((p) => (
                <Button
                  key={p}
                  variant={period === p ? 'default' : 'ghost'}
                  size="sm"
                  className="h-7 text-xs"
                  onClick={() => setPeriod(p)}
                >
                  {p === 'today' ? '今日' : p === 'week' ? '本周' : '本月'}
                </Button>
              ))}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {trends && trends.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={trends}>
                <defs>
                  <linearGradient id="callGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                <XAxis dataKey="date" fontSize={12} tickLine={false} axisLine={false} />
                <YAxis fontSize={12} tickLine={false} axisLine={false} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'hsl(var(--card))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: 'var(--radius)',
                    boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)',
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="calls"
                  stroke="hsl(var(--primary))"
                  strokeWidth={2}
                  fill="url(#callGradient)"
                />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex h-48 items-center justify-center text-muted-foreground">
              暂无趋势数据
            </div>
          )}
        </CardContent>
      </Card>

      {/* Top 10 Keys */}
      <Card>
        <CardHeader><CardTitle className="text-base">Top 10 卡密</CardTitle></CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-16">排名</TableHead>
                <TableHead>别名</TableHead>
                <TableHead className="text-right">调用数</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {topKeys?.slice(0, 10).map((item, i) => (
                <TableRow key={i} className="group">
                  <TableCell>
                    <span className={`inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold ${
                      i === 0 ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-950 dark:text-yellow-400' :
                      i === 1 ? 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300' :
                      i === 2 ? 'bg-orange-100 text-orange-700 dark:bg-orange-950 dark:text-orange-400' :
                      'bg-muted text-muted-foreground'
                    }`}>
                      {i + 1}
                    </span>
                  </TableCell>
                  <TableCell className="font-medium">{item.key_alias || '-'}</TableCell>
                  <TableCell className="text-right font-mono">{item.count.toLocaleString()}</TableCell>
                </TableRow>
              ))}
              {(!topKeys || topKeys.length === 0) && (
                <TableRow><TableCell colSpan={3} className="text-center text-muted-foreground py-8">暂无数据</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Recent Logs */}
      <Card>
        <CardHeader><CardTitle className="text-base">最近使用记录</CardTitle></CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Key 别名</TableHead>
                <TableHead>扣减数量</TableHead>
                <TableHead>IP</TableHead>
                <TableHead>时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {dash?.recent_logs?.slice(0, 20).map((log) => (
                <TableRow key={log.id}>
                  <TableCell className="font-medium">{log.key_alias}</TableCell>
                  <TableCell><Badge variant="outline">{log.amount}</Badge></TableCell>
                  <TableCell className="font-mono text-sm text-muted-foreground">{log.ip}</TableCell>
                  <TableCell className="text-muted-foreground">{new Date(log.created_at).toLocaleString('zh-CN')}</TableCell>
                </TableRow>
              ))}
              {(!dash?.recent_logs || dash.recent_logs.length === 0) && (
                <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground py-8">暂无使用记录</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
