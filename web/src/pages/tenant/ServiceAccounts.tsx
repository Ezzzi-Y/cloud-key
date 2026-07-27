import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listServiceAccounts, createServiceAccount, toggleServiceAccount, deleteServiceAccount } from '@/api/service-accounts'
import type { CreateServiceAccountResult } from '@/api/service-accounts'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { SkeletonTable } from '@/components/SkeletonTable'
import ServiceApiDocs from '@/components/ServiceApiDocs'
import { toast } from 'sonner'
import { Plus, RefreshCw, Trash2, MoreHorizontal, Power } from 'lucide-react'

export default function ServiceAccounts() {
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [rawKey, setRawKey] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)
  const queryClient = useQueryClient()

  const { data: accounts, isLoading } = useQuery({
    queryKey: ['tenant-service-accounts'],
    queryFn: () => listServiceAccounts().then((r) => (r.code === 0 ? r.data : [])),
  })

  const createMutation = useMutation({
    mutationFn: (name: string) => createServiceAccount(name),
    onSuccess: (res) => { if (res.code === 0) { setRawKey((res.data as CreateServiceAccountResult).raw_key); queryClient.invalidateQueries({ queryKey: ['tenant-service-accounts'] }); toast.success('服务账号创建成功') } else toast.error(res.message) },
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) => toggleServiceAccount(id, isActive),
    onSuccess: (res) => { if (res.code === 0) { queryClient.invalidateQueries({ queryKey: ['tenant-service-accounts'] }); toast.success('状态已切换') } else toast.error(res.message) },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteServiceAccount,
    onSuccess: (res) => { if (res.code === 0) { queryClient.invalidateQueries({ queryKey: ['tenant-service-accounts'] }); toast.success('已删除') } else toast.error(res.message); setDeleteTarget(null) },
  })

  const closeCreate = () => { setRawKey(''); setNewName(''); setCreateOpen(false) }

  return (
    <Tabs defaultValue="manage" className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">服务账号</h2>
        <div className="flex items-center gap-2">
          <TabsList>
            <TabsTrigger value="manage">服务账号管理</TabsTrigger>
            <TabsTrigger value="docs">API 接入文档</TabsTrigger>
          </TabsList>
          <Button variant="outline" size="icon" onClick={() => queryClient.invalidateQueries({ queryKey: ['tenant-service-accounts'] })}><RefreshCw className="h-4 w-4" /></Button>
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild><Button><Plus className="mr-2 h-4 w-4" />创建服务账号</Button></DialogTrigger>
            <DialogContent>
              {!rawKey ? (
                <>
                  <DialogHeader><DialogTitle>创建服务账号</DialogTitle><DialogDescription>密钥仅在创建时显示一次</DialogDescription></DialogHeader>
                  <div className="space-y-4"><div className="space-y-2"><Label>账号名称</Label><Input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="例如：my-service" /></div></div>
                  <DialogFooter>
                    <Button variant="outline" onClick={() => setCreateOpen(false)}>取消</Button>
                    <Button onClick={() => createMutation.mutate(newName.trim())} disabled={!newName.trim() || createMutation.isPending}>
                      {createMutation.isPending ? '创建中...' : '创建'}
                    </Button>
                  </DialogFooter>
                </>
              ) : (
                <>
                  <DialogHeader><DialogTitle>创建成功</DialogTitle><DialogDescription>请立即保存以下密钥</DialogDescription></DialogHeader>
                  <div className="rounded-lg bg-muted p-4"><code className="break-all text-sm font-bold">{rawKey}</code></div>
                  <DialogFooter><Button onClick={closeCreate}>我已保存，关闭</Button></DialogFooter>
                </>
              )}
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <TabsContent value="manage">
        <div className="space-y-6">

      {isLoading ? (
        <SkeletonTable rows={3} cols={4} />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader><TableRow><TableHead>名称</TableHead><TableHead>状态</TableHead><TableHead>创建时间</TableHead><TableHead className="w-12">操作</TableHead></TableRow></TableHeader>
              <TableBody>
                {accounts?.length === 0 ? <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground py-8">暂无服务账号</TableCell></TableRow> : null}
                {accounts?.map((sa) => (
                  <TableRow key={sa.id} className="group">
                    <TableCell className="font-medium">{sa.name}</TableCell>
                    <TableCell><Badge variant={sa.is_active ? 'success' : 'destructive'}>{sa.is_active ? '已启用' : '已禁用'}</Badge></TableCell>
                    <TableCell className="text-muted-foreground">{new Date(sa.created_at).toLocaleDateString('zh-CN')}</TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => toggleMutation.mutate({ id: sa.id, isActive: !sa.is_active })}>
                            <Power className="mr-2 h-4 w-4" />{sa.is_active ? '禁用' : '启用'}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setDeleteTarget(sa.id)}>
                            <Trash2 className="mr-2 h-4 w-4" />删除
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      <AlertDialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader><AlertDialogTitle>确认删除</AlertDialogTitle><AlertDialogDescription>此操作不可撤销</AlertDialogDescription></AlertDialogHeader>
          <AlertDialogFooter><AlertDialogCancel>取消</AlertDialogCancel><AlertDialogAction onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget)}>删除</AlertDialogAction></AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
        </div>
      </TabsContent>

      <TabsContent value="docs">
        <ServiceApiDocs />
      </TabsContent>
    </Tabs>
  )
}
