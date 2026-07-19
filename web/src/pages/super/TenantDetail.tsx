import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getTenant, updateTenant, resetTenantPassword, type UpdateTenantRequest } from '@/api/tenants'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useToast } from '@/hooks/use-toast'
import { ArrowLeft } from 'lucide-react'

interface Tenant {
  id: number
  name: string
  status: string
  expire_at: string
  key_count: number
  user_count: number
  key_prefix: string
  key_length: number
  key_suffix_length: number
}

export default function TenantDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [status, setStatus] = useState('active')
  const [expireAt, setExpireAt] = useState('')
  const [keyPrefix, setKeyPrefix] = useState('sk-')
  const [keyLength, setKeyLength] = useState(32)
  const [keySuffixLength, setKeySuffixLength] = useState(4)
  const [newPassword, setNewPassword] = useState<string | null>(null)

  const { data: tenant, isLoading } = useQuery({
    queryKey: ['tenant', id],
    queryFn: () => getTenant(Number(id)).then((r) => (r.code === 0 ? r.data : null)),
    enabled: !!id,
  })

  useEffect(() => {
    if (tenant) {
      const t = tenant as Tenant
      setName(t.name); setStatus(t.status); setExpireAt(t.expire_at || '')
      setKeyPrefix(t.key_prefix); setKeyLength(t.key_length); setKeySuffixLength(t.key_suffix_length)
    }
  }, [tenant])

  const updateMutation = useMutation({
    mutationFn: (data: UpdateTenantRequest) => updateTenant(Number(id), data),
    onSuccess: (res) => {
      if (res.code === 0) {
        toast({ title: '成功', description: '更新成功' })
        queryClient.invalidateQueries({ queryKey: ['tenant', id] })
        queryClient.invalidateQueries({ queryKey: ['super-tenants'] })
      } else { toast({ title: '错误', description: res.message, variant: 'destructive' }) }
    },
  })

  const resetMutation = useMutation({
    mutationFn: () => resetTenantPassword(Number(id)),
    onSuccess: (res) => {
      if (res.code === 0) { setNewPassword((res.data as any).new_password); toast({ title: '成功', description: '密码已重置' }) }
      else toast({ title: '错误', description: res.message, variant: 'destructive' })
    },
  })

  const handleSave = () => {
    updateMutation.mutate({
      name, status: status as 'active' | 'expired' | 'disabled', expire_at: expireAt || '', key_prefix: keyPrefix, key_length: keyLength, key_suffix_length: keySuffixLength,
    })
  }

  if (isLoading) return <div className="flex items-center justify-center py-12">加载中...</div>
  if (!tenant) return <div className="py-12 text-center text-muted-foreground">租户不存在</div>

  const t = tenant as Tenant

  return (
    <div className="space-y-6">
      <Button variant="ghost" onClick={() => navigate('/super/tenants')}>
        <ArrowLeft className="mr-2 h-4 w-4" />返回列表
      </Button>
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">租户详情：{t.name}</h2>
        <Badge variant={t.status === 'active' ? 'success' : t.status === 'expired' ? 'warning' : 'destructive'}>
          {t.status === 'active' ? '活跃' : t.status === 'expired' ? '已过期' : '已禁用'}
        </Badge>
      </div>
      <Card>
        <CardHeader><CardTitle>基本信息</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2"><Label>租户名称</Label><Input value={name} onChange={(e) => setName(e.target.value)} /></div>
            <div className="space-y-2"><Label>状态</Label>
              <Select value={status} onValueChange={setStatus}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">活跃</SelectItem>
                  <SelectItem value="expired">已过期</SelectItem>
                  <SelectItem value="disabled">已禁用</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2"><Label>到期时间（留空表示永不过期）</Label>
              <Input type="datetime-local" value={expireAt ? expireAt.replace(' ', 'T') : ''} onChange={(e) => setExpireAt(e.target.value.replace('T', ' '))} />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Key 配置</CardTitle></CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="space-y-2"><Label>Key 前缀</Label><Input value={keyPrefix} onChange={(e) => setKeyPrefix(e.target.value)} /></div>
            <div className="space-y-2"><Label>Key 长度</Label><Input type="number" value={keyLength} onChange={(e) => setKeyLength(Number(e.target.value))} /></div>
            <div className="space-y-2"><Label>后缀显示长度</Label><Input type="number" value={keySuffixLength} onChange={(e) => setKeySuffixLength(Number(e.target.value))} /></div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>统计信息</CardTitle></CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div><span className="text-muted-foreground">Key 总数：</span><span className="font-bold">{t.key_count}</span></div>
            <div><span className="text-muted-foreground">用户数：</span><span className="font-bold">{t.user_count}</span></div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>操作</CardTitle></CardHeader>
        <CardContent className="flex flex-wrap gap-4">
          <Button onClick={handleSave} disabled={updateMutation.isPending}>{updateMutation.isPending ? '保存中...' : '保存修改'}</Button>
          <Button variant="outline" onClick={() => resetMutation.mutate()} disabled={resetMutation.isPending}>{resetMutation.isPending ? '重置中...' : '重置管理员密码'}</Button>
          {newPassword && (
            <div className="flex items-center gap-2 rounded bg-muted px-4 py-2">
              <span className="text-sm">新密码：</span><code className="font-bold">{newPassword}</code>
              <Button variant="ghost" size="sm" onClick={() => setNewPassword(null)}>✕</Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
