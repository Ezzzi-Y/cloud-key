import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listKeys, createKey, disableKey, enableKey, deleteKey, exportKeysCSV, exportKeysJSON, updateKey } from '@/api/keys'
import type { KeyListParams, CreateKeyRequest, KeyBillingMode, KeyStatus, Key } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { SkeletonTable } from '@/components/SkeletonTable'
import { TablePagination } from '@/components/TablePagination'
import { toast } from 'sonner'
import { Plus, Download, RefreshCw, MoreHorizontal, Edit, Ban, CheckCircle, Trash2, FileDown } from 'lucide-react'

export default function KeyManagement() {
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState<KeyStatus | ''>('')
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)
  const [rawKey, setRawKey] = useState('')
  const [selectedKey, setSelectedKey] = useState<{ id: number; alias: string; remaining: number } | null>(null)
  const [newAlias, setNewAlias] = useState('')
  const [newBilling, setNewBilling] = useState<KeyBillingMode>('count')
  const [newAmount, setNewAmount] = useState(100)
  const [newExpireAt, setNewExpireAt] = useState('')
  const [newMaxUsage, setNewMaxUsage] = useState('')
  const [saving, setSaving] = useState(false)
  const queryClient = useQueryClient()

  const params: KeyListParams = { page, page_size: 20 }
  if (status) params.status = status as KeyStatus
  if (keyword) params.keyword = keyword

  const { data, isLoading } = useQuery({
    queryKey: ['tenant-keys', params],
    queryFn: () => listKeys(params).then((r) => (r.code === 0 ? r.data : { list: [], total: 0, page: 1, page_size: 20 })),
  })

  const createMutation = useMutation({
    mutationFn: (req: CreateKeyRequest) => createKey(req),
    onSuccess: (res) => {
      if (res.code === 0) { setRawKey((res.data as { raw_key: string; key: Key }).raw_key); toast.success('Key 创建成功'); queryClient.invalidateQueries({ queryKey: ['tenant-keys'] }) }
      else toast.error(res.message)
    },
  })

  const toggleDisableMutation = useMutation({
    mutationFn: disableKey,
    onSuccess: (res) => { if (res.code === 0) { toast.success('Key 已禁用'); queryClient.invalidateQueries({ queryKey: ['tenant-keys'] }) } else toast.error(res.message) },
  })
  const toggleEnableMutation = useMutation({
    mutationFn: enableKey,
    onSuccess: (res) => { if (res.code === 0) { toast.success('Key 已启用'); queryClient.invalidateQueries({ queryKey: ['tenant-keys'] }) } else toast.error(res.message) },
  })
  const deleteMutation = useMutation({
    mutationFn: deleteKey,
    onSuccess: (res) => { if (res.code === 0) { toast.success('Key 已删除'); queryClient.invalidateQueries({ queryKey: ['tenant-keys'] }) } else toast.error(res.message); setDeleteTarget(null) },
  })

  const handleCreate = () => {
    createMutation.mutate({ alias: newAlias, billing_mode: newBilling, initial_amount: newAmount, expire_at: newExpireAt || undefined, max_usage: newMaxUsage ? Number(newMaxUsage) : undefined })
  }

  const closeCreate = () => { setRawKey(''); setCreateOpen(false); setNewAlias(''); setNewAmount(100); setNewExpireAt(''); setNewMaxUsage('') }

  const handleExportCSV = async () => {
    try {
      const res = await exportKeysCSV()
      if (typeof res === 'string') {
        const blob = new Blob([res], { type: 'text/csv;charset=utf-8' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a'); a.href = url; a.download = 'keys.csv'; a.click()
        URL.revokeObjectURL(url)
      } else { toast.error((res as any)?.message || '导出失败') }
    } catch { toast.error('导出失败') }
  }

  const handleExportJSON = async () => {
    const res = await exportKeysJSON(); if (res.code === 0) { const blob = new Blob([JSON.stringify(res.data, null, 2)], { type: 'application/json' }); const url = URL.createObjectURL(blob); const a = document.createElement('a'); a.href = url; a.download = 'keys.json'; a.click(); URL.revokeObjectURL(url) }
  }

  const handleEdit = async () => {
    if (selectedKey) {
      setSaving(true)
      try {
        const res = await updateKey(selectedKey.id, { alias: selectedKey.alias, remaining_amount: selectedKey.remaining })
        if (res.code === 0) { toast.success('Key 已更新'); queryClient.invalidateQueries({ queryKey: ['tenant-keys'] }) }
        else toast.error(res.message)
        setEditOpen(false)
      } catch { toast.error('保存失败') }
      finally { setSaving(false) }
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">Key 管理</h2>
        <div className="flex gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm"><FileDown className="mr-2 h-4 w-4" />导出</Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={handleExportCSV}><Download className="mr-2 h-4 w-4" />导出 CSV</DropdownMenuItem>
              <DropdownMenuItem onClick={handleExportJSON}><Download className="mr-2 h-4 w-4" />导出 JSON</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <Button variant="outline" size="icon" onClick={() => queryClient.invalidateQueries({ queryKey: ['tenant-keys'] })}><RefreshCw className="h-4 w-4" /></Button>
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild><Button><Plus className="mr-2 h-4 w-4" />创建 Key</Button></DialogTrigger>
            <DialogContent>
              {!rawKey ? (
                <>
                  <DialogHeader><DialogTitle>创建新 Key</DialogTitle><DialogDescription>Key 明文仅在创建时显示一次，请妥善保存</DialogDescription></DialogHeader>
                  <div className="space-y-4">
                    <div className="space-y-2"><Label>别名</Label><Input value={newAlias} onChange={(e) => setNewAlias(e.target.value)} placeholder="可选" /></div>
                    <div className="space-y-2"><Label>计费模式</Label>
                      <Select value={newBilling} onValueChange={(v) => setNewBilling(v as KeyBillingMode)}>
                        <SelectTrigger><SelectValue /></SelectTrigger>
                        <SelectContent><SelectItem value="count">按次数</SelectItem><SelectItem value="credit">按 Credit</SelectItem></SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2"><Label>初始额度</Label><Input type="number" value={newAmount} onChange={(e) => setNewAmount(Number(e.target.value))} /></div>
                    <div className="space-y-2"><Label>过期时间（可选）</Label><Input type="datetime-local" value={newExpireAt} onChange={(e) => setNewExpireAt(e.target.value.replace('T', ' '))} /></div>
                    <div className="space-y-2"><Label>最大使用次数（可选）</Label><Input type="number" value={newMaxUsage} onChange={(e) => setNewMaxUsage(e.target.value)} /></div>
                  </div>
                  <DialogFooter>
                    <Button variant="outline" onClick={() => setCreateOpen(false)}>取消</Button>
                    <Button onClick={handleCreate} disabled={newAmount <= 0 || createMutation.isPending}>
                      {createMutation.isPending ? '创建中...' : '创建'}
                    </Button>
                  </DialogFooter>
                </>
              ) : (
                <>
                  <DialogHeader><DialogTitle>Key 创建成功</DialogTitle><DialogDescription>请立即保存以下 Key，关闭后将无法再次查看</DialogDescription></DialogHeader>
                  <div className="rounded-lg bg-muted p-4"><code className="break-all text-sm font-bold">{rawKey}</code></div>
                  <DialogFooter><Button onClick={closeCreate}>我已保存，关闭</Button></DialogFooter>
                </>
              )}
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="flex gap-4">
        <Input placeholder="搜索别名..." value={keyword} onChange={(e) => { setKeyword(e.target.value); setPage(1) }} className="max-w-xs" />
        <Select value={status} onValueChange={(v) => { setStatus(v as KeyStatus | ''); setPage(1) }}>
          <SelectTrigger className="w-32"><SelectValue placeholder="全部状态" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="">全部状态</SelectItem>
            <SelectItem value="unused">未使用</SelectItem>
            <SelectItem value="used">已用尽</SelectItem>
            <SelectItem value="disabled">已禁用</SelectItem>
            <SelectItem value="expired">已过期</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <SkeletonTable rows={5} cols={8} />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>别名</TableHead><TableHead>Key</TableHead><TableHead>计费模式</TableHead><TableHead>额度(初始/剩余)</TableHead>
                  <TableHead>状态</TableHead><TableHead>创建时间</TableHead><TableHead>最后使用</TableHead><TableHead className="w-12">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.list.length === 0 ? (
                  <TableRow><TableCell colSpan={8} className="text-center text-muted-foreground py-8">暂无 Key 数据</TableCell></TableRow>
                ) : (
                  data?.list.map((key: Key) => (
                    <TableRow key={key.id} className="group">
                      <TableCell className="font-medium">{key.alias || '-'}</TableCell>
                      <TableCell className="font-mono text-sm text-muted-foreground">{key.key_prefix}****{key.key_suffix}</TableCell>
                      <TableCell>{key.billing_mode === 'count' ? '按次数' : '按Credit'}</TableCell>
                      <TableCell>{key.initial_amount} / {key.remaining_amount}</TableCell>
                      <TableCell>
                        <Badge variant={key.status === 'unused' ? 'secondary' : key.status === 'used' ? 'outline' : key.status === 'disabled' ? 'destructive' : 'warning'}>
                          {key.status === 'unused' ? '可用' : key.status === 'used' ? '已用尽' : key.status === 'disabled' ? '已禁用' : '已过期'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-muted-foreground">{new Date(key.created_at).toLocaleDateString('zh-CN')}</TableCell>
                      <TableCell className="text-muted-foreground">{key.used_at ? new Date(key.used_at).toLocaleDateString('zh-CN') : '-'}</TableCell>
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => { setSelectedKey({ id: key.id, alias: key.alias, remaining: key.remaining_amount }); setEditOpen(true) }}>
                              <Edit className="mr-2 h-4 w-4" />编辑
                            </DropdownMenuItem>
                            {key.status === 'unused' && (
                              <DropdownMenuItem onClick={() => toggleDisableMutation.mutate(key.id)}>
                                <Ban className="mr-2 h-4 w-4" />禁用
                              </DropdownMenuItem>
                            )}
                            {key.status === 'disabled' && (
                              <DropdownMenuItem onClick={() => toggleEnableMutation.mutate(key.id)}>
                                <CheckCircle className="mr-2 h-4 w-4" />启用
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuSeparator />
                            <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setDeleteTarget(key.id)}>
                              <Trash2 className="mr-2 h-4 w-4" />删除
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {data && <TablePagination page={page} total={data.total} pageSize={20} onPageChange={setPage} />}

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>编辑 Key</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2"><Label>别名</Label><Input value={selectedKey?.alias || ''} onChange={(e) => setSelectedKey((p) => p ? { ...p, alias: e.target.value } : null)} /></div>
            <div className="space-y-2"><Label>剩余额度</Label><Input type="number" value={selectedKey?.remaining || 0} onChange={(e) => setSelectedKey((p) => p ? { ...p, remaining: Number(e.target.value) } : null)} /></div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>取消</Button>
            <Button onClick={handleEdit} disabled={saving}>{saving ? '保存中...' : '保存'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader><AlertDialogTitle>确认删除</AlertDialogTitle><AlertDialogDescription>此操作不可撤销，确定要删除此 Key 吗？</AlertDialogDescription></AlertDialogHeader>
          <AlertDialogFooter><AlertDialogCancel>取消</AlertDialogCancel><AlertDialogAction onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget)}>删除</AlertDialogAction></AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
