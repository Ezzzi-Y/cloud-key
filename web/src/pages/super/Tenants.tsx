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
import { useToast } from '@/hooks/use-toast'
import { Plus } from 'lucide-react'

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
  const { toast } = useToast()
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
        toast({ title: '成功', description: '租户创建成功' })
      } else {
        toast({ title: '错误', description: res.message, variant: 'destructive' })
      }
    },
    onError: () => toast({ title: '错误', description: '创建失败，请重试', variant: 'destructive' }),
  })

  const closeResult = () => { setResult(null); setDialogOpen(false); setNewName('') }

  if (isLoading) return <div className="flex items-center justify-center py-12">加载中...</div>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">租户管理</h2>
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
                <div className="space-y-2 rounded bg-muted p-4">
                  <p>用户名：<code className="font-bold">{result.username}</code></p>
                  <p>密码：<code className="font-bold">{result.password}</code></p>
                </div>
                <DialogFooter><Button onClick={closeResult}>我已保存，关闭</Button></DialogFooter>
              </>
            )}
          </DialogContent>
        </Dialog>
      </div>
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>到期时间</TableHead>
                <TableHead>Key 数量</TableHead>
                <TableHead>用户数</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(tenants as Tenant[])?.map((t) => (
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
                  <TableCell>{new Date(t.created_at).toLocaleDateString('zh-CN')}</TableCell>
                  <TableCell>
                    <Button variant="link" size="sm" onClick={() => navigate(`/super/tenants/${t.id}`)}>详情</Button>
                  </TableCell>
                </TableRow>
              ))}
              {(!tenants || (tenants as Tenant[]).length === 0) && (
                <TableRow><TableCell colSpan={7} className="text-center text-muted-foreground">暂无租户</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
