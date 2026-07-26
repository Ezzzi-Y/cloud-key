import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listUsageLogs, exportUsageLogs } from '@/api/logs'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { SkeletonTable } from '@/components/SkeletonTable'
import { TablePagination } from '@/components/TablePagination'
import { toast } from 'sonner'
import { Download, Info, RefreshCw } from 'lucide-react'

export default function UsageLogs() {
  const [page, setPage] = useState(1)
  const [alias, setAlias] = useState('')
  const [keySuffix, setKeySuffix] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['tenant-usage-logs', { page, alias, keySuffix, startTime, endTime }],
    queryFn: () => listUsageLogs({ page, page_size: 20, alias: alias || undefined, key_suffix: keySuffix || undefined, start_time: startTime ? startTime.replace('T', ' ') : undefined, end_time: endTime ? endTime.replace('T', ' ') : undefined })
      .then((r) => (r.code === 0 ? r.data : { list: [], total: 0, page: 1, page_size: 20 })),
  })

  const handleExport = async () => {
    try {
      const res = await exportUsageLogs({ alias: alias || undefined, key_suffix: keySuffix || undefined, start_time: startTime ? startTime.replace('T', ' ') : undefined, end_time: endTime ? endTime.replace('T', ' ') : undefined })
      if (res.code === 0) { const blob = new Blob([JSON.stringify(res.data, null, 2)], { type: 'application/json' }); const url = URL.createObjectURL(blob); const a = document.createElement('a'); a.href = url; a.download = 'usage-logs.json'; a.click(); URL.revokeObjectURL(url) }
    } catch { toast.error('导出失败') }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <h2 className="text-2xl font-bold tracking-tight">使用记录</h2>
        <span className="flex items-center gap-1 text-xs text-muted-foreground"><Info className="h-3 w-3" />数据可能存在一定延迟</span>
      </div>
      <div className="flex flex-wrap items-center gap-4">
        <Input placeholder="别名前缀搜索..." value={alias} onChange={(e) => { setAlias(e.target.value); setPage(1) }} className="max-w-[160px]" />
        <Input placeholder="后缀精准搜索..." value={keySuffix} onChange={(e) => { setKeySuffix(e.target.value); setPage(1) }} className="max-w-[160px]" />
        <Input type="datetime-local" value={startTime} onChange={(e) => { setStartTime(e.target.value); setPage(1) }} className="max-w-[200px]" />
        <Input type="datetime-local" value={endTime} onChange={(e) => { setEndTime(e.target.value); setPage(1) }} className="max-w-[200px]" />
        <div className="ml-auto flex gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()}><RefreshCw className="mr-2 h-4 w-4" />刷新</Button>
          <Button variant="outline" size="sm" onClick={handleExport}><Download className="mr-2 h-4 w-4" />导出 JSON</Button>
        </div>
      </div>

      {isLoading ? (
        <SkeletonTable rows={5} cols={6} />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow><TableHead>Key 别名</TableHead><TableHead>卡密后缀</TableHead><TableHead>扣减数量</TableHead><TableHead>IP</TableHead><TableHead>User-Agent</TableHead><TableHead>时间</TableHead></TableRow>
              </TableHeader>
              <TableBody>
                {data?.list.length === 0 ? <TableRow><TableCell colSpan={6} className="text-center text-muted-foreground py-8">暂无记录</TableCell></TableRow> : null}
                {data?.list.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="font-medium">{log.key_alias}</TableCell>
                    <TableCell className="font-mono text-sm text-muted-foreground">{log.key_suffix}</TableCell>
                    <TableCell><Badge variant="outline">{log.amount}</Badge></TableCell>
                    <TableCell className="font-mono text-sm text-muted-foreground">{log.ip}</TableCell>
                    <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">{log.user_agent}</TableCell>
                    <TableCell className="text-muted-foreground">{new Date(log.created_at).toLocaleString('zh-CN')}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {data && <TablePagination page={page} total={data.total} pageSize={20} onPageChange={setPage} />}
    </div>
  )
}
