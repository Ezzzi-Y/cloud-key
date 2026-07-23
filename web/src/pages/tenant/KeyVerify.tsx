import { useState } from 'react'
import { getKeyStatus, consumeKey } from '@/api/keys'
import type { KeyStatusResult, ConsumeKeyResult } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import { Search, Minus } from 'lucide-react'

export default function KeyVerify() {
  const [rawKey, setRawKey] = useState('')
  const [checking, setChecking] = useState(false)
  const [result, setResult] = useState<KeyStatusResult | null>(null)
  const [consumeAmount, setConsumeAmount] = useState(1)
  const [consuming, setConsuming] = useState(false)
  const [consumeResult, setConsumeResult] = useState<ConsumeKeyResult | null>(null)

  const handleCheck = async () => {
    if (!rawKey.trim()) { toast.error('请输入卡密'); return }
    setChecking(true); setResult(null); setConsumeResult(null)
    try {
      const res = await getKeyStatus(rawKey.trim())
      if (res.code === 0) { setResult(res.data as KeyStatusResult) }
      else { toast.error(res.message) }
    } catch { toast.error('请求失败') }
    finally { setChecking(false) }
  }

  const handleConsume = async () => {
    if (!rawKey.trim() || consumeAmount <= 0) return
    setConsuming(true); setConsumeResult(null)
    try {
      const res = await consumeKey({ key: rawKey.trim(), amount: consumeAmount })
      if (res.code === 0) {
        const d = res.data as ConsumeKeyResult; setConsumeResult(d)
        if (result) setResult({ ...result, remaining_amount: d.remaining_amount, status: d.status })
        toast.success(`扣减成功，剩余额度：${d.remaining_amount}`)
      } else { toast.error(res.message) }
    } catch { toast.error('请求失败') }
    finally { setConsuming(false) }
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Key 校验与扣减</h2>
      <Card>
        <CardHeader><CardTitle>输入卡密</CardTitle><CardDescription>粘贴用户提供的 Key 完整明文进行查询和扣减</CardDescription></CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <Input value={rawKey} onChange={(e) => setRawKey(e.target.value)} placeholder="例如：sk-a1b2c3d4..." className="flex-1 font-mono" onKeyDown={(e) => e.key === 'Enter' && handleCheck()} />
            <Button onClick={handleCheck} disabled={checking}><Search className="mr-2 h-4 w-4" />{checking ? '校验中...' : '校验'}</Button>
          </div>
        </CardContent>
      </Card>

      {result && (
        <Card>
          <CardHeader><CardTitle>校验结果</CardTitle></CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              <div><div className="text-sm text-muted-foreground">别名</div><div className="font-medium">{result.alias || '-'}</div></div>
              <div><div className="text-sm text-muted-foreground">剩余额度</div><div className="font-medium">{result.remaining_amount}</div></div>
              <div><div className="text-sm text-muted-foreground">状态</div>
                <Badge variant={result.status === 'active' ? 'secondary' : result.status === 'exhausted' ? 'outline' : result.status === 'disabled' ? 'destructive' : 'warning'}>
                  {result.status === 'active' ? '可用' : result.status === 'exhausted' ? '已用尽' : result.status === 'disabled' ? '已禁用' : '已过期'}
                </Badge>
              </div>
              <div><div className="text-sm text-muted-foreground">创建时间</div><div className="font-medium">{new Date(result.created_at).toLocaleString('zh-CN')}</div></div>
              <div><div className="text-sm text-muted-foreground">最后使用</div><div className="font-medium">{result.used_at ? new Date(result.used_at).toLocaleString('zh-CN') : '未使用'}</div></div>
            </div>
          </CardContent>
        </Card>
      )}

      {result && result.status === 'active' && (
        <Card>
          <CardHeader><CardTitle>扣减操作</CardTitle><CardDescription>扣减此 Key 的额度，扣减后立即生效</CardDescription></CardHeader>
          <CardContent>
            <div className="flex items-end gap-4">
              <div className="space-y-2"><Label>扣减数量</Label><Input type="number" value={consumeAmount} onChange={(e) => setConsumeAmount(Number(e.target.value))} min={1} className="w-32" /></div>
              <Button onClick={handleConsume} disabled={consuming || consumeAmount <= 0}><Minus className="mr-2 h-4 w-4" />{consuming ? '扣减中...' : '扣减'}</Button>
            </div>
            {consumeResult && (
              <div className="mt-4 rounded border p-4">
                <div className="grid gap-2 md:grid-cols-3">
                  <div><div className="text-sm text-muted-foreground">剩余额度</div><div className="text-lg font-bold">{consumeResult.remaining_amount}</div></div>
                  <div><div className="text-sm text-muted-foreground">状态</div><Badge variant={consumeResult.used_up ? 'outline' : 'secondary'}>{consumeResult.used_up ? '已用尽' : '可用'}</Badge></div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
