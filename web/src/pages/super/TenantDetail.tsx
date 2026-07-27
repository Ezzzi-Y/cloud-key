import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getTenant, updateTenant, resetTenantPassword, type UpdateTenantRequest } from '@/api/tenants'
import type { TenantListItem } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { DateTimePicker } from '@/components/ui/date-time-picker'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from 'sonner'
import { Skeleton } from '@/components/ui/skeleton'
import { ArrowLeft } from 'lucide-react'

export default function TenantDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [status, setStatus] = useState('active')
  const [expireAt, setExpireAt] = useState('')
  const [newPassword, setNewPassword] = useState<string | null>(null)

  const { data: tenant, isLoading } = useQuery({
    queryKey: ['tenant', id],
    queryFn: () => getTenant(Number(id)).then((r) => (r.code === 0 ? (r.data as TenantListItem) : null)),
    enabled: !!id,
  })

  useEffect(() => {
    if (tenant) {
      setName(tenant.name)
      setStatus(tenant.status)
      if (tenant.expire_at) {
        // Backend returns RFC3339 e.g. "2025-01-15T14:30:00+08:00"
        // datetime-local input needs "YYYY-MM-DDTHH:mm" (no timezone)
        const dt = new Date(tenant.expire_at)
        const yyyy = dt.getFullYear()
        const mm = String(dt.getMonth() + 1).padStart(2, '0')
        const dd = String(dt.getDate()).padStart(2, '0')
        const hh = String(dt.getHours()).padStart(2, '0')
        const mi = String(dt.getMinutes()).padStart(2, '0')
        setExpireAt(`${yyyy}-${mm}-${dd}T${hh}:${mi}`)
      } else {
        setExpireAt('')
      }
    }
  }, [tenant])

  const updateMutation = useMutation({
    mutationFn: (data: UpdateTenantRequest) => updateTenant(Number(id), data),
    onSuccess: (res) => {
      if (res.code === 0) {
        toast.success('更新成功')
        queryClient.invalidateQueries({ queryKey: ['tenant', id] })
        queryClient.invalidateQueries({ queryKey: ['super-tenants'] })
      } else { toast.error(res.message) }
    },
  })

  const resetMutation = useMutation({
    mutationFn: () => resetTenantPassword(Number(id)),
    onSuccess: (res) => {
      if (res.code === 0) { setNewPassword((res.data as { new_password: string }).new_password); toast.success('密码已重置') }
      else toast.error(res.message)
    },
  })

  const handleSave = () => {
    // expireAt is "YYYY-MM-DDTHH:mm", convert to "YYYY-MM-DD HH:MM:SS" for backend
    const formatted = expireAt ? expireAt.replace('T', ' ') + ':00' : ''
    updateMutation.mutate({
      name, status: status as 'active' | 'expired' | 'disabled', expire_at: formatted,
    })
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-24" />
        <Skeleton className="h-8 w-64" />
        <div className="grid gap-4 md:grid-cols-2"><Skeleton className="h-10" /><Skeleton className="h-10" /><Skeleton className="h-10" /></div>
      </div>
    )
  }
  if (!tenant) return <div className="py-12 text-center text-muted-foreground">租户不存在</div>

  return (
    <div className="space-y-6">
      <Button variant="ghost" onClick={() => navigate('/super/tenants')}>
        <ArrowLeft className="mr-2 h-4 w-4" />返回列表
      </Button>
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">租户详情：{tenant.name}</h2>
        <Badge variant={tenant.status === 'active' ? 'success' : tenant.status === 'expired' ? 'warning' : 'destructive'}>
          {tenant.status === 'active' ? '活跃' : tenant.status === 'expired' ? '已过期' : '已禁用'}
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
              <DateTimePicker value={expireAt} onChange={setExpireAt} placeholder="选择到期时间" clearable />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>统计信息</CardTitle></CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div><span className="text-muted-foreground">Key 总数：</span><span className="font-bold">{tenant.key_count}</span></div>
            <div><span className="text-muted-foreground">用户数：</span><span className="font-bold">{tenant.user_count}</span></div>
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
