import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { login, verify2FA, setupTOTPInit, confirmTOTPInit } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useToast } from '@/hooks/use-toast'
import type { ApiResponse, LoginStep1Data, LoginResponse } from '@/types'

type Step = 'form' | 'totp' | 'setup'

export default function LoginPage() {
  const [step, setStep] = useState<Step>('form')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [userId, setUserId] = useState(0)
  const [totpCode, setTotpCode] = useState('')
  const [qrUrl, setQrUrl] = useState('')
  const [secret, setSecret] = useState('')
  const [loading, setLoading] = useState(false)
  const { login: authLogin } = useAuth()
  const navigate = useNavigate()
  const { toast } = useToast()

  const handleApiError = (res: ApiResponse) => {
    toast({ title: '错误', description: res.message || '操作失败', variant: 'destructive' })
  }

  const handleLogin = async (e: FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await login(username, password)
      if (res.code !== 0) { handleApiError(res); return }
      const data = res.data as LoginStep1Data
      setUserId(data.user_id)
      if (data.require_totp) {
        setStep('totp')
      } else if (data.need_setup) {
        await handleInitSetup(data.user_id)
      }
    } catch {
      toast({ title: '错误', description: '登录失败，请检查网络连接', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  const handleVerifyTOTP = async (e: FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await verify2FA(userId, totpCode)
      if (res.code !== 0) { handleApiError(res); return }
      const data = res.data as LoginResponse
      authLogin(data.token, data.role, data.tenant_id, data.username)
      navigate(data.role === 'super_admin' ? '/super/' : '/tenant/')
    } catch {
      toast({ title: '错误', description: '验证失败', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  const handleInitSetup = async (uid: number) => {
    setLoading(true)
    try {
      const res = await setupTOTPInit(uid)
      if (res.code !== 0) { handleApiError(res); return }
      setQrUrl((res.data as { url: string }).url)
      setSecret((res.data as { secret: string }).secret)
      setStep('setup')
    } catch {
      toast({ title: '错误', description: 'TOTP 设置初始化失败', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  const handleConfirmSetup = async (e: FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await confirmTOTPInit(userId, totpCode)
      if (res.code !== 0) { handleApiError(res); return }
      const data = res.data as LoginResponse
      authLogin(data.token, data.role, data.tenant_id, data.username)
      navigate(data.role === 'super_admin' ? '/super/' : '/tenant/')
    } catch {
      toast({ title: '错误', description: 'TOTP 设置确认失败', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">CloudKey 管理后台</CardTitle>
          <CardDescription>
            {step === 'form' && '请输入管理员账号密码'}
            {step === 'totp' && '请输入 TOTP 验证码'}
            {step === 'setup' && '首次登录 — 请设置 TOTP 两步验证'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {step === 'form' && (
            <form onSubmit={handleLogin} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="username">用户名</Label>
                <Input id="username" value={username} onChange={(e) => setUsername(e.target.value)} placeholder="请输入用户名" required autoFocus />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">密码</Label>
                <Input id="password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="请输入密码" required />
              </div>
              <Button type="submit" className="w-full" disabled={loading}>{loading ? '登录中...' : '登录'}</Button>
            </form>
          )}

          {step === 'totp' && (
            <form onSubmit={handleVerifyTOTP} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="totp">TOTP 验证码</Label>
                <Input id="totp" value={totpCode} onChange={(e) => setTotpCode(e.target.value)} placeholder="请输入 6 位验证码" maxLength={6} required autoFocus />
              </div>
              <Button type="submit" className="w-full" disabled={loading}>{loading ? '验证中...' : '验证'}</Button>
              <Button type="button" variant="ghost" className="w-full" onClick={() => setStep('form')}>返回</Button>
            </form>
          )}

          {step === 'setup' && (
            <form onSubmit={handleConfirmSetup} className="space-y-4">
              {qrUrl && (
                <div className="flex flex-col items-center gap-4">
                  <p className="text-sm text-muted-foreground">请使用 Google Authenticator 或类似应用扫描下方二维码</p>
                  <img src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrUrl)}`} alt="TOTP QR" className="rounded border" />
                  <p className="text-xs text-muted-foreground break-all">或手动输入密钥：<code className="select-all">{secret}</code></p>
                </div>
              )}
              <div className="space-y-2">
                <Label htmlFor="setup-totp">验证码</Label>
                <Input id="setup-totp" value={totpCode} onChange={(e) => setTotpCode(e.target.value)} placeholder="请输入 6 位验证码以确认设置" maxLength={6} required autoFocus />
              </div>
              <Button type="submit" className="w-full" disabled={loading}>{loading ? '确认中...' : '确认并登录'}</Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
