import { useState, type FormEvent, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { login, verify2FA, setupTOTPInit, confirmTOTPInit } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import type { ApiResponse, LoginStep1Data, LoginResponse } from '@/types'
import { Loader2, ArrowLeft, KeyRound } from 'lucide-react'

type Step = 'form' | 'totp' | 'setup'

const Logo = (props: React.SVGProps<SVGSVGElement>) => (
  <svg fill="currentColor" viewBox="0 0 40 48" width="48" height="58" {...props}>
    <clipPath id="a"><path d="m0 0h40v48h-40z" /></clipPath>
    <g clipPath="url(#a)">
      <path d="m25.0887 5.05386-3.933-1.05386-3.3145 12.3696-2.9923-11.16736-3.9331 1.05386 3.233 12.0655-8.05262-8.0526-2.87919 2.8792 8.83271 8.8328-10.99975-2.9474-1.05385625 3.933 12.01860625 3.2204c-.1376-.5935-.2104-1.2119-.2104-1.8473 0-4.4976 3.646-8.1436 8.1437-8.1436 4.4976 0 8.1436 3.646 8.1436 8.1436 0 .6313-.0719 1.2459-.2078 1.8359l10.9227 2.9267 1.0538-3.933-12.0664-3.2332 11.0005-2.9476-1.0539-3.933-12.0659 3.233 8.0526-8.0526-2.8792-2.87916-8.7102 8.71026z" />
      <path d="m27.8723 26.2214c-.3372 1.4256-1.0491 2.7063-2.0259 3.7324l7.913 7.9131 2.8792-2.8792z" />
      <path d="m25.7665 30.0366c-.9886 1.0097-2.2379 1.7632-3.6389 2.1515l2.8794 10.746 3.933-1.0539z" />
      <path d="m21.9807 32.2274c-.65.1671-1.3313.2559-2.0334.2559-.7522 0-1.4806-.102-2.1721-.2929l-2.882 10.7558 3.933 1.0538z" />
      <path d="m17.6361 32.1507c-1.3796-.4076-2.6067-1.1707-3.5751-2.1833l-7.9325 7.9325 2.87919 2.8792z" />
      <path d="m13.9956 29.8973c-.9518-1.019-1.6451-2.2826-1.9751-3.6862l-10.95836 2.9363 1.05385 3.933z" />
    </g>
  </svg>
)

export default function LoginPage() {
  const [step, setStep] = useState<Step>('form')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [userId, setUserId] = useState(0)
  const [preAuthToken, setPreAuthToken] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [qrUrl, setQrUrl] = useState('')
  const [secret, setSecret] = useState('')
  const [loading, setLoading] = useState(false)
  const { login: authLogin, isAuthenticated, role } = useAuth()
  const navigate = useNavigate()

  // 已登录用户重定向到对应仪表盘
  useEffect(() => {
    if (isAuthenticated && role) {
      navigate(role === 'super_admin' ? '/super/' : '/tenant/', { replace: true })
    }
  }, [isAuthenticated, role, navigate])

  const handleApiError = (res: ApiResponse) => {
    toast.error(res.message || '操作失败')
  }

  const handleLogin = async (e: FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await login(username, password)
      if (res.code !== 0) { handleApiError(res); return }
      const data = res.data as LoginStep1Data
      setUserId(data.user_id)
      setPreAuthToken(data.pre_auth_token)
      if (data.require_totp) {
        setStep('totp')
      } else if (data.need_setup) {
        await handleInitSetup(data.user_id)
      }
    } catch (err: unknown) {
      const msg = (err as { message?: string })?.message
      toast.error(msg || '登录失败，请检查网络连接')
    } finally {
      setLoading(false)
    }
  }

  const handleVerifyTOTP = async (e: FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await verify2FA(userId, totpCode, preAuthToken)
      if (res.code !== 0) { handleApiError(res); return }
      const data = res.data as LoginResponse
      authLogin(data.token, data.role, data.tenant_id, data.username, data.tenant_status, data.tenant_expire_at)
      navigate(data.role === 'super_admin' ? '/super/' : '/tenant/')
    } catch {
      toast.error('验证失败')
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
      toast.error('验证器绑定初始化失败')
    } finally {
      setLoading(false)
    }
  }

  const handleConfirmSetup = async (e: FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await confirmTOTPInit(userId, totpCode, preAuthToken)
      if (res.code !== 0) { handleApiError(res); return }
      const data = res.data as LoginResponse
      authLogin(data.token, data.role, data.tenant_id, data.username, data.tenant_status, data.tenant_expire_at)
      navigate(data.role === 'super_admin' ? '/super/' : '/tenant/')
    } catch {
      toast.error('验证器绑定确认失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <Card className="w-full max-w-sm rounded-3xl px-6 py-10 pt-14 shadow-lg">
        <CardContent>
          <div className="flex flex-col items-center space-y-8">
            {/* Logo */}
            <Logo className="text-foreground" />

            {/* Title area */}
            <div className="space-y-2 text-center">
              {step === 'form' && (
                <>
                  <h1 className="text-3xl font-semibold text-foreground">欢迎回来</h1>
                  <p className="text-sm text-muted-foreground">请输入您的管理员账号和密码</p>
                </>
              )}
              {step === 'totp' && (
                <>
                  <h1 className="text-3xl font-semibold text-foreground">安全验证</h1>
                  <p className="text-sm text-muted-foreground">
                    请提供您的一次性密码验证器应用提供的密码
                  </p>
                </>
              )}
              {step === 'setup' && (
                <>
                  <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                    <KeyRound className="h-6 w-6 text-primary" />
                  </div>
                  <h1 className="text-3xl font-semibold text-foreground">绑定验证器</h1>
                  <p className="text-sm text-muted-foreground">
                    首次登录，请使用一次性密码验证器应用扫描下方二维码完成绑定
                  </p>
                </>
              )}
            </div>

            {/* Step: Login form */}
            {step === 'form' && (
              <form onSubmit={handleLogin} className="w-full space-y-4">
                <Input
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="用户名"
                  className="w-full rounded-xl"
                  required
                  autoFocus
                />
                <Input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="密码"
                  className="w-full rounded-xl"
                  required
                />
                <Button type="submit" className="w-full rounded-xl" size="lg" disabled={loading}>
                  {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  {loading ? '登录中...' : '登录'}
                </Button>
              </form>
            )}

            {/* Step: TOTP verify */}
            {step === 'totp' && (
              <form onSubmit={handleVerifyTOTP} className="w-full space-y-4">
                <Input
                  id="totp"
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value)}
                  placeholder="6位验证码"
                  maxLength={6}
                  className="w-full rounded-xl text-center text-lg tracking-[0.5em] font-mono"
                  required
                  autoFocus
                />
                <Button type="submit" className="w-full rounded-xl" size="lg" disabled={loading}>
                  {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  {loading ? '验证中...' : '验证'}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  className="w-full text-sm text-muted-foreground"
                  onClick={() => { setStep('form'); setTotpCode(''); setPreAuthToken('') }}
                >
                  <ArrowLeft className="mr-1 h-3 w-3" />
                  返回登录
                </Button>
              </form>
            )}

            {/* Step: TOTP setup */}
            {step === 'setup' && (
              <form onSubmit={handleConfirmSetup} className="w-full space-y-4">
                {qrUrl && (
                  <div className="flex flex-col items-center gap-4">
                    <div className="rounded-2xl border-2 border-border bg-white p-3 shadow-sm">
                      <img
                        src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrUrl)}`}
                        alt="扫码绑定验证器"
                        className="rounded-lg"
                      />
                    </div>
                    <div className="w-full rounded-xl bg-muted/50 p-3 text-center">
                      <p className="text-xs text-muted-foreground mb-1">或手动输入密钥：</p>
                      <code className="text-sm font-mono font-bold select-all break-all">{secret}</code>
                    </div>
                  </div>
                )}

                <Separator />

                <div className="space-y-2">
                  <Label htmlFor="setup-totp" className="text-center block">输入验证码以确认绑定</Label>
                  <Input
                    id="setup-totp"
                    value={totpCode}
                    onChange={(e) => setTotpCode(e.target.value)}
                    placeholder="6 位验证码"
                    maxLength={6}
                    className="w-full rounded-xl text-center text-lg tracking-[0.5em] font-mono"
                    required
                    autoFocus
                  />
                </div>
                <Button type="submit" className="w-full rounded-xl" size="lg" disabled={loading}>
                  {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  {loading ? '绑定中...' : '确认并登录'}
                </Button>
              </form>
            )}

            {/* Footer */}
            <p className="w-11/12 text-center text-xs text-muted-foreground">
              CloudKey
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
