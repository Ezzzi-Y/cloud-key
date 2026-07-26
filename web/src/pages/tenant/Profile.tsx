import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProfile, changePassword, setupTOTP, confirmTOTP } from '@/api/auth'
import { listLoginLogs } from '@/api/logs'
import { getKeyConfig, updateKeyConfig, type KeyConfig as KeyConfigType, type UpdateKeyConfigRequest } from '@/api/keys'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import { useAuth } from '@/hooks/useAuth'

export default function TenantProfile() {
  const { username } = useAuth()
  const role = 'tenant'
  const queryClient = useQueryClient()

  const { data: profile } = useQuery({
    queryKey: ['tenant-profile'],
    queryFn: () => getProfile(role).then((r) => (r.code === 0 ? r.data : null)),
  })

  const [oldPass, setOldPass] = useState(''); const [newPass, setNewPass] = useState('')
  const passMutation = useMutation({
    mutationFn: () => changePassword(role, oldPass, newPass),
    onSuccess: (res) => { if (res.code === 0) { toast.success('密码修改成功'); setOldPass(''); setNewPass('') } else toast.error(res.message) },
  })

  const [totpCode, setTotpCode] = useState(''); const [qrUrl, setQrUrl] = useState(''); const [totpSecret, setTotpSecret] = useState('')
  const setupMutation = useMutation({
    mutationFn: () => setupTOTP(role),
    onSuccess: (res) => { if (res.code === 0) { const d = res.data as { secret: string; url: string }; setTotpSecret(d.secret); setQrUrl(d.url) } else toast.error(res.message) },
  })
  const confirmMutation = useMutation({
    mutationFn: () => confirmTOTP(role, totpCode),
    onSuccess: (res) => { if (res.code === 0) { toast.success('验证器重新绑定成功'); setQrUrl(''); setTotpSecret(''); setTotpCode('') } else toast.error(res.message) },
  })

  const [loginPage, setLoginPage] = useState(1)
  const { data: logsData } = useQuery({
    queryKey: ['tenant-login-logs', loginPage],
    queryFn: () => listLoginLogs(role, loginPage).then((r) => r.code === 0 ? r.data : { list: [], total: 0, page: 1, page_size: 20 }),
  })

  // ========== Key 配置 ==========
  const [keyPrefix, setKeyPrefix] = useState('')
  const [keyLength, setKeyLength] = useState(32)
  const [keySuffixLength, setKeySuffixLength] = useState(4)
  const [prefixError, setPrefixError] = useState('')
  const [lengthError, setLengthError] = useState('')
  const [suffixError, setSuffixError] = useState('')

  const { data: keyConfig } = useQuery({
    queryKey: ['tenant-key-config'],
    queryFn: () => getKeyConfig().then((r) => (r.code === 0 ? (r.data as KeyConfigType) : null)),
  })

  useEffect(() => {
    if (keyConfig) {
      setKeyPrefix(keyConfig.key_prefix)
      setKeyLength(keyConfig.key_length)
      setKeySuffixLength(keyConfig.key_suffix_length)
    }
  }, [keyConfig])

  const keyConfigMutation = useMutation({
    mutationFn: (data: UpdateKeyConfigRequest) => updateKeyConfig(data),
    onSuccess: (res) => {
      if (res.code === 0) { toast.success('Key 配置已更新'); queryClient.invalidateQueries({ queryKey: ['tenant-key-config'] }) }
      else toast.error(res.message)
    },
  })

  const validateKeyConfig = (): boolean => {
    let valid = true
    setPrefixError(''); setLengthError(''); setSuffixError('')

    if (!keyPrefix || keyPrefix.length < 1 || keyPrefix.length > 10) {
      setPrefixError('前缀长度需在 1~10 之间'); valid = false
    } else if (!/^[a-zA-Z0-9_]+-$/.test(keyPrefix)) {
      setPrefixError('必须以 - 结尾，仅允许字母、数字、下划线'); valid = false
    }
    if (keyLength < 8 || keyLength > 32) {
      setLengthError('Key 长度需在 8~32 之间'); valid = false
    }
    if (keySuffixLength < 4) {
      setSuffixError('后缀显示长度不能小于 4'); valid = false
    } else if (keySuffixLength > keyLength) {
      setSuffixError('后缀显示长度不能超过 Key 长度'); valid = false
    }
    return valid
  }

  const handleKeyConfigSave = () => {
    if (!validateKeyConfig()) return
    keyConfigMutation.mutate({ key_prefix: keyPrefix, key_length: keyLength, key_suffix_length: keySuffixLength })
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">个人设置</h2>
      <Tabs defaultValue="profile">
        <TabsList>
          <TabsTrigger value="profile">资料 & 密码</TabsTrigger>
          <TabsTrigger value="key-config">Key 配置</TabsTrigger>
          <TabsTrigger value="totp">验证器管理</TabsTrigger>
          <TabsTrigger value="history">登录日志</TabsTrigger>
        </TabsList>
        <TabsContent value="profile">
          <Card>
            <CardHeader><CardTitle>个人信息</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <div><span className="text-muted-foreground">用户名：</span>{username}</div>
                <div><span className="text-muted-foreground">角色：</span>租户管理员</div>
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
        <TabsContent value="key-config">
          <Card>
            <CardHeader>
              <CardTitle>卡密生成规则</CardTitle>
              <CardDescription>配置新建卡密时的生成格式。平台仅存储末尾几位用于显示和识别卡密，修改仅影响后续创建的卡密。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid gap-6 md:grid-cols-3">
                <div className="space-y-2">
                  <Label>Key 前缀</Label>
                  <Input value={keyPrefix} onChange={(e) => { setKeyPrefix(e.target.value); setPrefixError('') }} placeholder="sk-" maxLength={10} />
                  <p className="text-xs text-muted-foreground">1~10 字符，必须以 - 结尾</p>
                  {prefixError && <p className="text-xs text-destructive">{prefixError}</p>}
                </div>
                <div className="space-y-2">
                  <Label>Key 长度</Label>
                  <Input type="number" value={keyLength} onChange={(e) => { setKeyLength(Number(e.target.value)); setLengthError('') }} min={8} max={32} />
                  <p className="text-xs text-muted-foreground">随机部分长度，范围 8~32</p>
                  {lengthError && <p className="text-xs text-destructive">{lengthError}</p>}
                </div>
                <div className="space-y-2">
                  <Label>后缀显示长度</Label>
                  <Input type="number" value={keySuffixLength} onChange={(e) => { setKeySuffixLength(Number(e.target.value)); setSuffixError('') }} min={4} />
                  <p className="text-xs text-muted-foreground">平台存储用于显示的位数，平台不完整存储Key。最少4位</p>
                  {suffixError && <p className="text-xs text-destructive">{suffixError}</p>}
                </div>
              </div>
              <div className="rounded-md bg-muted p-4">
                <p className="text-sm text-muted-foreground mb-1">生成示例：</p>
                <code className="text-sm font-mono">
                  {keyPrefix}
                  {'x'.repeat(Math.min(keyLength, 8))}...
                  <span className="text-primary font-bold"> (末 {keySuffixLength} 位显示)</span>
                </code>
              </div>
              <Button onClick={handleKeyConfigSave} disabled={keyConfigMutation.isPending}>
                {keyConfigMutation.isPending ? '保存中...' : '保存配置'}
              </Button>
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
                  {logsData?.list.map((l) => (
                    <TableRow key={l.id}>
                      <TableCell>{l.ip}</TableCell>
                      <TableCell className="max-w-xs truncate">{l.user_agent}</TableCell>
                      <TableCell><Badge variant={l.status === 'success' ? 'success' : 'destructive'}>{l.status === 'success' ? '成功' : '失败'}</Badge></TableCell>
                      <TableCell>{new Date(l.created_at).toLocaleString('zh-CN')}</TableCell>
                    </TableRow>
                  ))}
                  {(!logsData || logsData.list.length === 0) && (
                    <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground">暂无登录记录</TableCell></TableRow>
                  )}
                </TableBody>
              </Table>
              {logsData && logsData.total > 20 && (
                <div className="mt-4 flex justify-center gap-2">
                  <Button variant="outline" size="sm" disabled={loginPage <= 1} onClick={() => setLoginPage((p) => p - 1)}>上一页</Button>
                  <span className="flex items-center text-sm">{loginPage} / {Math.ceil(logsData.total / 20)}</span>
                  <Button variant="outline" size="sm" disabled={loginPage >= Math.ceil(logsData.total / 20)} onClick={() => setLoginPage((p) => p + 1)}>下一页</Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
