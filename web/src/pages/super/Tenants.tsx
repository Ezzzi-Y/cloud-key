import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { createTenant, listTenants } from '@/api/tenants'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { SkeletonTable } from '@/components/SkeletonTable'
import { toast } from 'sonner'
import { Plus, ExternalLink } from 'lucide-react'

interface Tenant {
  id: number
  name: string
  status: string
  expire_at: string
  key_count: number
  user_count: number
  created_at: string
}

export default function TenantsPage() {
  const [newName, setNewName] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [result, setResult] = useState<{ username: string; password: string } | null>(null)
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: tenants, isLoading } = useQuery({
    queryKey: ['super-tenants'],
    queryFn: () => listTenants().then((r) => (r.code === 0 ? r.data : [])),
  })

  const createMutation = useMutation({
    mutationFn: (name: string) => createTenant(name),
    onSuccess: (res) => {
      if (res.code === 0) {
        setResult({ username: (res.data as any).admin_username, password: (res.data as any).admin_password })
        queryClient.invalidateQueries({ queryKey: ['super-tenants'] })
        toast.success('租户创建成功')
      } else { toast.error(res.message) }
    },
    onError: () => toast.error('创建失败，请重试'),
  })

  const closeResult = () => { setResult(null); setDialogOpen(false); setNewName('') }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">租户管理</h2>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button><Plus className="mr-2 h-4 w-4" />创建租户</Button>
          </DialogTrigger>
          <DialogContent>
            {!result ? (
              <>
                <DialogHeader>
                  <DialogTitle>创建新租户</DialogTitle>
                  <DialogDescription>请输入租户名称，系统将自动生成管理员账号和密码</DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <Label htmlFor="name">租户名称</Label>
                    <Input id="name" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="例如：myapp" />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setDialogOpen(false)}>取消</Button>
                  <Button onClick={() => createMutation.mutate(newName.trim())} disabled={!newName.trim() || createMutation.isPending}>
                    {createMutation.isPending ? '创建中...' : '创建'}
                  </Button>
                </DialogFooter>
              </>
            ) : (
              <>
                <DialogHeader>
                  <DialogTitle>租户创建成功</DialogTitle>
                  <DialogDescription>请保存以下管理员账号信息，关闭后将无法再次查看密码</DialogDescription>
                </DialogHeader>
                <div className="space-y-2 rounded-lg bg-muted p-4">
                  <p>用户名：<code className="font-bold">{result.username}</code></p>
                  <p>密码：<code className="font-bold">{result.password}</code></p>
                </div>
                <DialogFooter><Button onClick={closeResult}>我已保存，关闭</Button></DialogFooter>
              </>
            )}
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <SkeletonTable rows={5} cols={7} />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>到期时间</TableHead>
                  <TableHead className="text-right">Key 数量</TableHead>
                  <TableHead className="text-right">用户数</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(tenants as Tenant[])?.map((t) => (
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
                    <TableCell className="text-muted-foreground">{new Date(t.created_at).toLocaleDateString('zh-CN')}</TableCell>
                    <TableCell className="text-right">
                      <Button variant="link" size="sm" onClick={() => navigate(`/super/tenants/${t.id}`)} className="gap-1">
                        详情 <ExternalLink className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
                {(!tenants || (tenants as Tenant[]).length === 0) && (
                  <TableRow><TableCell colSpan={7} className="text-center text-muted-foreground py-8">暂无租户</TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
