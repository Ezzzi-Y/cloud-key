import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listUsageLogs, exportUsageLogs } from '@/api/logs'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useToast } from '@/hooks/use-toast'
import { Download } from 'lucide-react'

export default function UsageLogs() {
  const [page, setPage] = useState(1)
  const [keyAlias, setKeyAlias] = useState('')
  const [ip, setIp] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const { toast } = useToast()

  const { data, isLoading } = useQuery({
    queryKey: ['tenant-usage-logs', { page, keyAlias, ip, startTime, endTime }],
    queryFn: () => listUsageLogs({ page, page_size: 20, key_alias: keyAlias || undefined, ip: ip || undefined, start_time: startTime ? startTime.replace('T', ' ') : undefined, end_time: endTime ? endTime.replace('T', ' ') : undefined })
      .then((r) => (r.code === 0 ? r.data : { list: [], total: 0, page: 1, page_size: 20 })),
  })

  const handleExport = async () => {
    try {
      const res = await exportUsageLogs({ key_alias: keyAlias || undefined, ip: ip || undefined, start_time: startTime ? startTime.replace('T', ' ') : undefined, end_time: endTime ? endTime.replace('T', ' ') : undefined })
      if (res.code === 0) { const blob = new Blob([JSON.stringify(res.data, null, 2)], { type: 'application/json' }); const url = URL.createObjectURL(blob); const a = document.createElement('a'); a.href = url; a.download = 'usage-logs.json'; a.click(); URL.revokeObjectURL(url) }
    } catch { toast({ title: '错误', description: '导出失败', variant: 'destructive' }) }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">使用记录</h2>
        <Button variant="outline" size="sm" onClick={handleExport}><Download className="mr-2 h-4 w-4" />导出 JSON</Button>
      </div>
      <div className="flex flex-wrap gap-4">
        <Input placeholder="Key 别名" value={keyAlias} onChange={(e) => { setKeyAlias(e.target.value); setPage(1) }} className="max-w-[160px]" />
        <Input placeholder="IP 地址" value={ip} onChange={(e) => { setIp(e.target.value); setPage(1) }} className="max-w-[160px]" />
        <Input type="datetime-local" value={startTime} onChange={(e) => { setStartTime(e.target.value); setPage(1) }} className="max-w-[200px]" />
        <Input type="datetime-local" value={endTime} onChange={(e) => { setEndTime(e.target.value); setPage(1) }} className="max-w-[200px]" />
      </div>
      <Card><CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow><TableHead>Key 别名</TableHead><TableHead>扣减数量</TableHead><TableHead>IP</TableHead><TableHead>User-Agent</TableHead><TableHead>请求路径</TableHead><TableHead>时间</TableHead></TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? <TableRow><TableCell colSpan={6} className="text-center">加载中...</TableCell></TableRow> : null}
            {!isLoading && data?.list.length === 0 ? <TableRow><TableCell colSpan={6} className="text-center text-muted-foreground">暂无记录</TableCell></TableRow> : null}
            {data?.list.map((log) => (
              <TableRow key={log.id}>
                <TableCell>{log.key_alias}</TableCell><TableCell>{log.amount}</TableCell>
                <TableCell className="font-mono text-sm">{log.ip}</TableCell>
                <TableCell className="max-w-[200px] truncate text-sm">{log.user_agent}</TableCell>
                <TableCell className="font-mono text-sm">{log.request_path}</TableCell>
                <TableCell>{new Date(log.created_at).toLocaleString('zh-CN')}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent></Card>
      {data && data.total > 20 && (
        <div className="flex justify-center gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>上一页</Button>
          <span className="flex items-center text-sm">{page} / {Math.ceil(data.total / 20)}</span>
          <Button variant="outline" size="sm" disabled={page >= Math.ceil(data.total / 20)} onClick={() => setPage((p) => p + 1)}>下一页</Button>
        </div>
      )}
    </div>
  )
}
