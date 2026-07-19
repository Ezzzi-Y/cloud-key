import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listServiceAccounts, createServiceAccount, toggleServiceAccount, deleteServiceAccount } from '@/api/service-accounts'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useToast } from '@/hooks/use-toast'
import { Plus, RefreshCw, Trash2 } from 'lucide-react'

export default function ServiceAccounts() {
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [rawKey, setRawKey] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)
  const { toast } = useToast()
  const queryClient = useQueryClient()

  const { data: accounts, isLoading } = useQuery({
    queryKey: ['tenant-service-accounts'],
    queryFn: () => listServiceAccounts().then((r) => (r.code === 0 ? r.data : [])),
  })

  const createMutation = useMutation({
    mutationFn: (name: string) => createServiceAccount(name),
    onSuccess: (res) => { if (res.code === 0) { setRawKey((res.data as { account: unknown; raw_key: string }).raw_key); queryClient.invalidateQueries({ queryKey: ['tenant-service-accounts'] }); toast({ title: '成功', description: '服务账号创建成功' }) } else toast({ title: '错误', description: res.message, variant: 'destructive' }) },
  })

  const toggleMutation = useMutation({
    mutationFn: toggleServiceAccount,
    onSuccess: (res) => { if (res.code === 0) { queryClient.invalidateQueries({ queryKey: ['tenant-service-accounts'] }); toast({ title: '成功', description: '状态已切换' }) } else toast({ title: '错误', description: res.message, variant: 'destructive' }) },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteServiceAccount,
    onSuccess: (res) => { if (res.code === 0) { queryClient.invalidateQueries({ queryKey: ['tenant-service-accounts'] }); toast({ title: '成功', description: '已删除' }) } else toast({ title: '错误', description: res.message, variant: 'destructive' }); setDeleteTarget(null) },
  })

  const closeCreate = () => { setRawKey(''); setNewName(''); setCreateOpen(false) }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">服务账号</h2>
        <div className="flex gap-2">
          <Button variant="outline" size="icon" onClick={() => queryClient.invalidateQueries({ queryKey: ['tenant-service-accounts'] })}><RefreshCw className="h-4 w-4" /></Button>
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild><Button><Plus className="mr-2 h-4 w-4" />创建服务账号</Button></DialogTrigger>
            <DialogContent>
              {!rawKey ? (
                <>
                  <DialogHeader><DialogTitle>创建服务账号</DialogTitle><DialogDescription>密钥仅在创建时显示一次</DialogDescription></DialogHeader>
                  <div className="space-y-4"><div className="space-y-2"><Label>账号名称</Label><Input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="例如：my-service" /></div></div>
                  <DialogFooter><Button variant="outline" onClick={() => setCreateOpen(false)}>取消</Button><Button onClick={() => createMutation.mutate(newName.trim())} disabled={!newName.trim()}>创建</Button></DialogFooter>
                </>
              ) : (
                <>
                  <DialogHeader><DialogTitle>创建成功</DialogTitle><DialogDescription>请立即保存以下密钥</DialogDescription></DialogHeader>
                  <div className="rounded bg-muted p-4"><code className="break-all text-sm font-bold">{rawKey}</code></div>
                  <DialogFooter><Button onClick={closeCreate}>我已保存，关闭</Button></DialogFooter>
                </>
              )}
            </DialogContent>
          </Dialog>
        </div>
      </div>
      <Card><CardContent className="p-0">
        <Table>
          <TableHeader><TableRow><TableHead>名称</TableHead><TableHead>状态</TableHead><TableHead>创建时间</TableHead><TableHead>操作</TableHead></TableRow></TableHeader>
          <TableBody>
            {isLoading ? <TableRow><TableCell colSpan={4} className="text-center">加载中...</TableCell></TableRow> : null}
            {!isLoading && accounts?.length === 0 ? <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground">暂无服务账号</TableCell></TableRow> : null}
            {accounts?.map((sa) => (
              <TableRow key={sa.id}>
                <TableCell className="font-medium">{sa.name}</TableCell>
                <TableCell><Badge variant={sa.is_active ? 'success' : 'destructive'}>{sa.is_active ? '已启用' : '已禁用'}</Badge></TableCell>
                <TableCell>{new Date(sa.created_at).toLocaleDateString('zh-CN')}</TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    <Button variant="outline" size="sm" onClick={() => toggleMutation.mutate(sa.id)}>{sa.is_active ? '禁用' : '启用'}</Button>
                    <Button variant="ghost" size="icon" onClick={() => setDeleteTarget(sa.id)}><Trash2 className="h-4 w-4 text-destructive" /></Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent></Card>
      <AlertDialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader><AlertDialogTitle>确认删除</AlertDialogTitle><AlertDialogDescription>此操作不可撤销</AlertDialogDescription></AlertDialogHeader>
          <AlertDialogFooter><AlertDialogCancel>取消</AlertDialogCancel><AlertDialogAction onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget)}>删除</AlertDialogAction></AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
