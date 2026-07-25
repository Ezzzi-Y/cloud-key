import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listBalanceLogs, exportBalanceLogs } from '@/api/logs'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { SkeletonTable } from '@/components/SkeletonTable'
import { TablePagination } from '@/components/TablePagination'
import { toast } from 'sonner'
import { Download, Info, RefreshCw } from 'lucide-react'

export default function BalanceLogs() {
  const [page, setPage] = useState(1)
  const [keyId, setKeyId] = useState('')
  const [operator, setOperator] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['tenant-balance-logs', { page, keyId, operator, startTime, endTime }],
    queryFn: () =>
      listBalanceLogs({
        page,
        page_size: 20,
        key_id: keyId ? Number(keyId) : undefined,
        operator: operator || undefined,
        start_time: startTime ? startTime.replace('T', ' ') : undefined,
        end_time: endTime ? endTime.replace('T', ' ') : undefined,
      }).then((r) => (r.code === 0 ? r.data : { list: [], total: 0, page: 1, page_size: 20 })),
  })

  const handleExport = async () => {
    try {
      const res = await exportBalanceLogs({
        key_id: keyId ? Number(keyId) : undefined,
        operator: operator || undefined,
        start_time: startTime ? startTime.replace('T', ' ') : undefined,
        end_time: endTime ? endTime.replace('T', ' ') : undefined,
      })
      if (res.code === 0) {
        const blob = new Blob([JSON.stringify(res.data, null, 2)], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = 'balance-logs.json'
        a.click()
        URL.revokeObjectURL(url)
      }
    } catch {
      toast.error('导出失败')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <h2 className="text-2xl font-bold tracking-tight">额度调整记录</h2>
        <span className="flex items-center gap-1 text-xs text-muted-foreground"><Info className="h-3 w-3" />数据可能存在一定延迟</span>
      </div>
      <div className="flex flex-wrap items-center gap-4">
        <Input placeholder="卡密 ID" type="number" value={keyId} onChange={(e) => { setKeyId(e.target.value); setPage(1) }} className="max-w-[120px]" />
        <Input placeholder="操作人" value={operator} onChange={(e) => { setOperator(e.target.value); setPage(1) }} className="max-w-[160px]" />
        <Input type="datetime-local" value={startTime} onChange={(e) => { setStartTime(e.target.value); setPage(1) }} className="max-w-[200px]" />
        <Input type="datetime-local" value={endTime} onChange={(e) => { setEndTime(e.target.value); setPage(1) }} className="max-w-[200px]" />
        <div className="ml-auto flex gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()}><RefreshCw className="mr-2 h-4 w-4" />刷新</Button>
          <Button variant="outline" size="sm" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />导出 JSON
          </Button>
        </div>
      </div>

      {isLoading ? (
        <SkeletonTable rows={5} cols={7} />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Key 别名</TableHead>
                  <TableHead>变动量</TableHead>
                  <TableHead>调整前</TableHead>
                  <TableHead>调整后</TableHead>
                  <TableHead>操作人</TableHead>
                  <TableHead>备注</TableHead>
                  <TableHead>时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.list.length === 0 ? (
                  <TableRow><TableCell colSpan={7} className="text-center text-muted-foreground py-8">暂无记录</TableCell></TableRow>
                ) : (
                  data?.list.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="font-medium">{log.key_alias}</TableCell>
                      <TableCell>
                        <Badge variant={log.delta > 0 ? 'success' : 'destructive'}>
                          {log.delta > 0 ? `+${log.delta}` : log.delta}
                        </Badge>
                      </TableCell>
                      <TableCell>{log.before_amount}</TableCell>
                      <TableCell>{log.after_amount}</TableCell>
                      <TableCell className="text-muted-foreground">{log.operator}</TableCell>
                      <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">{log.remark || '-'}</TableCell>
                      <TableCell className="text-muted-foreground">{new Date(log.created_at).toLocaleString('zh-CN')}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {data && <TablePagination page={page} total={data.total} pageSize={20} onPageChange={setPage} />}
    </div>
  )
}
