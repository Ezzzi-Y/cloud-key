import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { changePassword, setupTOTP, confirmTOTP } from '@/api/auth'
import { listLoginLogs } from '@/api/logs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import { useAuth } from '@/hooks/useAuth'

interface LoginLog {
  id: number
  ip: string
  user_agent: string
  status: string
  created_at: string
}

interface LogsData {
  list: LoginLog[]
  total: number
  page: number
  page_size: number
}

export default function SuperProfile() {
  const { username } = useAuth()
  const role = 'super'

  const [oldPass, setOldPass] = useState(''); const [newPass, setNewPass] = useState('')
  const passMutation = useMutation({
    mutationFn: () => changePassword(role, oldPass, newPass),
    onSuccess: (res) => {
      if (res.code === 0) { toast.success('密码修改成功'); setOldPass(''); setNewPass('') }
      else toast.error(res.message)
    },
  })

  const [totpCode, setTotpCode] = useState(''); const [qrUrl, setQrUrl] = useState(''); const [totpSecret, setTotpSecret] = useState('')
  const setupMutation = useMutation({
    mutationFn: () => setupTOTP(role),
    onSuccess: (res) => { if (res.code === 0) { const d = res.data as any; setTotpSecret(d.secret); setQrUrl(d.url) } else toast.error(res.message) },
  })
  const confirmMutation = useMutation({
    mutationFn: () => confirmTOTP(role, totpCode),
    onSuccess: (res) => {
      if (res.code === 0) { toast.success('验证器重新绑定成功'); setQrUrl(''); setTotpSecret(''); setTotpCode('') }
      else toast.error(res.message)
    },
  })

  const [loginPage, setLoginPage] = useState(1)
  const { data: logsData } = useQuery({
    queryKey: ['super-login-logs', loginPage],
    queryFn: () => listLoginLogs(role, loginPage).then((r) => r.code === 0 ? r.data : { list: [], total: 0, page: 1, page_size: 20 }),
  })

  const typedLogsData = logsData as LogsData | undefined

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">个人设置</h2>
      <Tabs defaultValue="profile">
        <TabsList>
          <TabsTrigger value="profile">资料 & 密码</TabsTrigger>
          <TabsTrigger value="totp">验证器管理</TabsTrigger>
          <TabsTrigger value="history">登录日志</TabsTrigger>
        </TabsList>
        <TabsContent value="profile">
          <Card>
            <CardHeader><CardTitle>个人信息</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <div><span className="text-muted-foreground">用户名：</span>{username}</div>
                <div><span className="text-muted-foreground">角色：</span>超级管理员</div>
              </div>
              <div className="border-t pt-4">
                <h3 className="mb-4 font-semibold">修改密码</h3>
                <div className="space-y-3 max-w-sm">
                  <div className="space-y-1"><Label>旧密码</Label><Input type="password" value={oldPass} onChange={(e) => setOldPass(e.target.value)} /></div>
                  <div className="space-y-1"><Label>新密码</Label><Input type="password" value={newPass} onChange={(e) => setNewPass(e.target.value)} /></div>
                  <Button onClick={() => passMutation.mutate()} disabled={!oldPass || !newPass}>修改密码</Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="totp">
          <Card>
            <CardHeader><CardTitle>验证器管理</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground">重新生成验证器密钥后，需要使用一次性密码验证器应用重新扫码绑定。</p>
              <Button onClick={() => setupMutation.mutate()} disabled={setupMutation.isPending}>重新生成验证器密钥</Button>
              {qrUrl && (
                <div className="space-y-4">
                  <img src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrUrl)}`} alt="TOTP QR" className="rounded border" />
                  <p className="text-xs text-muted-foreground break-all">密钥：{totpSecret}</p>
                  <div className="flex gap-2 max-w-sm">
                    <Input value={totpCode} onChange={(e) => setTotpCode(e.target.value)} placeholder="输入验证码确认" maxLength={6} />
                    <Button onClick={() => confirmMutation.mutate()} disabled={!totpCode}>确认</Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="history">
          <Card>
            <CardHeader><CardTitle>登录日志</CardTitle></CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow><TableHead>IP</TableHead><TableHead>User-Agent</TableHead><TableHead>状态</TableHead><TableHead>时间</TableHead></TableRow>
                </TableHeader>
                <TableBody>
                  {typedLogsData?.list.map((l) => (
                    <TableRow key={l.id}>
                      <TableCell>{l.ip}</TableCell>
                      <TableCell className="max-w-xs truncate">{l.user_agent}</TableCell>
                      <TableCell><Badge variant={l.status === 'success' ? 'success' : 'destructive'}>{l.status === 'success' ? '成功' : '失败'}</Badge></TableCell>
                      <TableCell>{new Date(l.created_at).toLocaleString('zh-CN')}</TableCell>
                    </TableRow>
                  ))}
                  {(!typedLogsData || typedLogsData.list.length === 0) && (
                    <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground">暂无登录记录</TableCell></TableRow>
                  )}
                </TableBody>
              </Table>
              {typedLogsData && typedLogsData.total > 20 && (
                <div className="mt-4 flex justify-center gap-2">
                  <Button variant="outline" size="sm" disabled={loginPage <= 1} onClick={() => setLoginPage((p) => p - 1)}>上一页</Button>
                  <span className="flex items-center text-sm">{loginPage} / {Math.ceil(typedLogsData.total / 20)}</span>
                  <Button variant="outline" size="sm" disabled={loginPage >= Math.ceil(typedLogsData.total / 20)} onClick={() => setLoginPage((p) => p + 1)}>下一页</Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
