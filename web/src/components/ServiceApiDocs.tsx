import { useState, useRef, useCallback, useEffect } from 'react'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { cn } from '@/lib/utils'
import { Copy, Check, ChevronDown } from 'lucide-react'

// ── Sidebar nav ──────────────────────────────────────────────────────

const SIDEBAR_GROUPS = [
  {
    label: '基础',
    items: [
      { id: 'quickstart', label: '快速开始' },
      { id: 'auth', label: '认证方式' },
    ],
  },
  {
    label: '接口',
    items: [
      { id: 'keys', label: '卡密管理' },
      { id: 'status', label: '状态管理' },
      { id: 'consume', label: '消费扣减' },
      { id: 'adjust', label: '额度调整' },
      { id: 'export', label: '导出与查询' },
    ],
  },
  {
    label: '进阶',
    items: [
      { id: 'idempotency', label: '幂等机制' },
      { id: 'errors', label: '错误码' },
    ],
  },
]

// ── Endpoint data ────────────────────────────────────────────────────

const METHOD_STYLES: Record<string, string> = {
  GET:   'bg-sky-50 text-sky-700 border-sky-200 dark:bg-sky-950 dark:text-sky-300 dark:border-sky-800',
  POST:  'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300 dark:border-emerald-800',
  PATCH: 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950 dark:text-amber-300 dark:border-amber-800',
  DELETE:'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950 dark:text-rose-300 dark:border-rose-800',
}

type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE'

interface Param {
  name: string
  type: string
  required?: boolean
  desc: string
}

interface Endpoint {
  method: HttpMethod
  path: string
  desc: string
  requestParams?: Param[]   // query params
  requestBody?: Param[]     // JSON body
  response?: Param[]        // response fields
}

interface EndpointGroup {
  title: string
  description: string
  endpoints: Endpoint[]
}

const GROUPS: Record<string, EndpointGroup> = {
  keys: {
    title: '卡密管理',
    description: '管理卡密的完整生命周期：创建、查询、更新和删除。',
    endpoints: [
      {
        method: 'POST', path: '/service/keys', desc: '创建新卡密',
        requestBody: [
          { name: 'alias', type: 'string', required: true, desc: '卡密别名，用于标识用途' },
          { name: 'remaining_amount', type: 'int64', required: true, desc: '初始额度' },
        ],
        response: [
          { name: 'id', type: 'int64', desc: '卡密 ID' },
          { name: 'raw_key', type: 'string', desc: '明文卡密值（仅创建时返回）' },
          { name: 'alias', type: 'string', desc: '别名' },
          { name: 'key_suffix', type: 'string', desc: '卡密后缀（用于展示）' },
          { name: 'remaining_amount', type: 'int64', desc: '剩余额度' },
          { name: 'status', type: 'string', desc: '状态：active / exhausted / disabled' },
        ],
      },
      {
        method: 'GET', path: '/service/keys', desc: '分页查询卡密列表',
        requestParams: [
          { name: 'page', type: 'int', desc: '页码，默认 1' },
          { name: 'page_size', type: 'int', desc: '每页数量，默认 20' },
          { name: 'status', type: 'string', desc: '状态过滤：active / exhausted / disabled / expired' },
          { name: 'alias', type: 'string', desc: '别名前缀搜索' },
          { name: 'key_suffix', type: 'string', desc: '后缀精准匹配' },
        ],
        response: [
          { name: 'list', type: 'array', desc: '卡密列表' },
          { name: 'total', type: 'int64', desc: '总数量' },
          { name: 'page', type: 'int', desc: '当前页码' },
          { name: 'page_size', type: 'int', desc: '每页数量' },
        ],
      },
      {
        method: 'GET', path: '/service/keys/:id', desc: '查询单个卡密详情',
        response: [
          { name: 'id', type: 'int64', desc: '卡密 ID' },
          { name: 'alias', type: 'string', desc: '别名' },
          { name: 'key_suffix', type: 'string', desc: '卡密后缀' },
          { name: 'remaining_amount', type: 'int64', desc: '剩余额度' },
          { name: 'status', type: 'string', desc: '当前状态' },
          { name: 'created_at', type: 'string', desc: '创建时间' },
        ],
      },
      {
        method: 'PATCH', path: '/service/keys/:id', desc: '更新卡密信息',
        requestBody: [
          { name: 'alias', type: 'string', desc: '新别名' },
        ],
        response: [],
      },
      {
        method: 'DELETE', path: '/service/keys/:id', desc: '删除卡密',
        response: [],
      },
    ],
  },
  status: {
    title: '状态管理',
    description: '启用、禁用卡密，或按卡密值直接查询状态。',
    endpoints: [
      {
        method: 'GET', path: '/service/keys/status', desc: '按卡密值查询状态',
        requestParams: [
          { name: 'sk', type: 'string', required: true, desc: '完整卡密值' },
        ],
        response: [
          { name: 'id', type: 'int64', desc: '卡密 ID' },
          { name: 'alias', type: 'string', desc: '别名' },
          { name: 'status', type: 'string', desc: '当前状态' },
          { name: 'remaining_amount', type: 'int64', desc: '剩余额度' },
        ],
      },
      {
        method: 'PATCH', path: '/service/keys/:id/enable', desc: '启用卡密',
        response: [],
      },
      {
        method: 'PATCH', path: '/service/keys/:id/disable', desc: '禁用卡密',
        response: [],
      },
    ],
  },
  consume: {
    title: '消费扣减',
    description: '扣减卡密额度。支持幂等，通过 X-Request-Id 防重复扣减。',
    endpoints: [
      {
        method: 'POST', path: '/service/keys/consume', desc: '扣减卡密额度',
        requestBody: [
          { name: 'key', type: 'string', required: true, desc: '完整卡密值' },
          { name: 'amount', type: 'int64', required: true, desc: '扣减数量（必须 > 0）' },
        ],
        response: [
          { name: 'key_id', type: 'int64', desc: '卡密 ID' },
          { name: 'key_alias', type: 'string', desc: '卡密别名' },
          { name: 'key_suffix', type: 'string', desc: '卡密后缀' },
          { name: 'consumed', type: 'int64', desc: '本次扣减量' },
          { name: 'remaining_amount', type: 'int64', desc: '扣减后剩余额度' },
        ],
      },
    ],
  },
  adjust: {
    title: '额度调整',
    description: '通过 delta 值增减卡密额度，正数增加、负数减少。所有变动记录在额度流转日志中。',
    endpoints: [
      {
        method: 'POST', path: '/service/keys/:id/adjust-balance', desc: '增减卡密额度',
        requestBody: [
          { name: 'delta', type: 'int64', required: true, desc: '调整量，正数增加、负数减少，不能为 0' },
          { name: 'remark', type: 'string', desc: '备注说明' },
        ],
        response: [
          { name: 'key_id', type: 'int64', desc: '卡密 ID' },
          { name: 'before_amount', type: 'int64', desc: '调整前额度' },
          { name: 'after_amount', type: 'int64', desc: '调整后额度' },
        ],
      },
    ],
  },
  export: {
    title: '导出与查询',
    description: '导出卡密数据、查询操作结果和额度流转日志。',
    endpoints: [
      {
        method: 'GET', path: '/service/keys/export', desc: '导出全部卡密（文本格式）',
        response: [
          { name: 'list', type: 'array', desc: '卡密文本列表' },
        ],
      },
      {
        method: 'GET', path: '/service/keys/export/json', desc: '导出全部卡密（JSON 格式）',
        response: [
          { name: 'list', type: 'array', desc: '卡密 JSON 数组，含 id / alias / key_suffix / remaining_amount / status' },
        ],
      },
      {
        method: 'GET', path: '/service/consume-result', desc: '按 request_id 查询操作结果',
        requestParams: [
          { name: 'request_id', type: 'string', required: true, desc: '请求 ID（UUID）' },
        ],
        response: [
          { name: 'source', type: 'string', desc: '数据来源：cache / usage_log / balance_log' },
          { name: 'request_id', type: 'string', desc: '请求 ID' },
          { name: 'data', type: 'object', desc: '操作结果（字段因操作类型而异）' },
        ],
      },
      {
        method: 'GET', path: '/service/balance-logs', desc: '查询额度流转日志',
        requestParams: [
          { name: 'page', type: 'int', desc: '页码，默认 1' },
          { name: 'page_size', type: 'int', desc: '每页数量，默认 20' },
        ],
        response: [
          { name: 'list', type: 'array', desc: '日志列表' },
          { name: 'total', type: 'int64', desc: '总数量' },
        ],
      },
      {
        method: 'GET', path: '/service/balance-logs/export', desc: '导出额度流转日志',
        response: [
          { name: 'list', type: 'array', desc: '日志列表' },
        ],
      },
    ],
  },
}

// ── Code examples ────────────────────────────────────────────────────

const LANGS = ['curl', 'java', 'js'] as const
type Lang = (typeof LANGS)[number]
const LANG_LABELS: Record<Lang, string> = { curl: 'cURL', java: 'Java', js: 'JavaScript' }

interface CodeExample { curl: string; java: string; js: string }

const EXAMPLES: Record<string, CodeExample> = {
  createKey: {
    curl: `curl -X POST https://your-domain.com/api/service/keys \\
  -H "X-Service-Key: sk_your_service_key" \\
  -H "Content-Type: application/json" \\
  -d '{"alias":"my-key","remaining_amount":100}'`,
    java: `import com.github.ezzzi_y.CloudKey;

CloudKey ck = new CloudKey("svc_your_service_key");

var key = ck.keys().create("my-key", 100L);
System.out.println("Key: " + key.getRawKey());`,
    js: `const res = await fetch('/api/service/keys', {
  method: 'POST',
  headers: {
    'X-Service-Key': 'sk_your_service_key',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ alias: 'my-key', remaining_amount: 100 }),
});
const { data } = await res.json();
console.log('Key:', data.raw_key);`,
  },
  consume: {
    curl: `curl -X POST https://your-domain.com/api/service/keys/consume \\
  -H "X-Service-Key: sk_your_service_key" \\
  -H "X-Request-Id: $(uuidgen)" \\
  -H "Content-Type: application/json" \\
  -d '{"key":"ck_abc123","amount":10}'`,
    java: `var result = ck.keys().consume("ck_abc123", 10L);
System.out.println("Remaining: " + result.getRemainingAmount());`,
    js: `const res = await fetch('/api/service/keys/consume', {
  method: 'POST',
  headers: {
    'X-Service-Key': 'sk_your_service_key',
    'X-Request-Id': crypto.randomUUID(),
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ key: 'ck_abc123', amount: 10 }),
});
const { data } = await res.json();
console.log('Remaining:', data.remaining_amount);`,
  },
  adjustBalance: {
    curl: `# 增加额度
curl -X POST https://your-domain.com/api/service/keys/1/adjust-balance \\
  -H "X-Service-Key: sk_your_service_key" \\
  -H "X-Request-Id: $(uuidgen)" \\
  -H "Content-Type: application/json" \\
  -d '{"delta":50,"remark":"充值"}'

# 减少额度
curl -X POST https://your-domain.com/api/service/keys/1/adjust-balance \\
  -H "X-Service-Key: sk_your_service_key" \\
  -H "X-Request-Id: $(uuidgen)" \\
  -H "Content-Type: application/json" \\
  -d '{"delta":-20,"remark":"扣款"}'`,
    java: `var adj = ck.keys().adjustBalance(1, 50L, "充值");
System.out.println("Before: " + adj.getBeforeAmount());
System.out.println("After: " + adj.getAfterAmount());`,
    js: `const res = await fetch('/api/service/keys/1/adjust-balance', {
  method: 'POST',
  headers: {
    'X-Service-Key': 'sk_your_service_key',
    'X-Request-Id': crypto.randomUUID(),
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ delta: 50, remark: '充值' }),
});
const { data } = await res.json();
console.log('After:', data.after_amount);`,
  },
  getKeyStatus: {
    curl: `curl "https://your-domain.com/api/service/keys/status?sk=ck_abc123" \\
  -H "X-Service-Key: sk_your_service_key"`,
    java: `var result = api.serviceGetKeyStatus("ck_abc123");
System.out.println("Status: " + result.getStatus());
System.out.println("Remaining: " + result.getRemainingAmount());`,
    js: `const res = await fetch(
  '/api/service/keys/status?sk=ck_abc123',
  { headers: { 'X-Service-Key': 'sk_your_service_key' } }
);
const { data } = await res.json();
console.log('Status:', data.status);`,
  },
  consumeResult: {
    curl: `curl "https://your-domain.com/api/service/consume-result?request_id=550e8400-..." \\
  -H "X-Service-Key: sk_your_service_key"`,
    java: `var result = api.serviceGetConsumeResult("550e8400-...");
System.out.println("Source: " + result.getSource());
System.out.println("Amount: " + result.getAmount());`,
    js: `const res = await fetch(
  '/api/service/consume-result?request_id=550e8400-...',
  { headers: { 'X-Service-Key': 'sk_your_service_key' } }
);
const { data } = await res.json();
console.log('Source:', data.source, 'Amount:', data.amount);`,
  },
}

// ── useActiveSection ─────────────────────────────────────────────────

function useActiveSection(ids: string[]) {
  const [active, setActive] = useState(ids[0])

  useEffect(() => {
    const observers: IntersectionObserver[] = []
    const visible = new Map<string, number>()

    ids.forEach((id) => {
      const el = document.getElementById(id)
      if (!el) return
      const obs = new IntersectionObserver(
        ([entry]) => {
          if (entry.isIntersecting) visible.set(id, entry.intersectionRatio)
          else visible.delete(id)
          let best = '', bestR = 0
          visible.forEach((r, k) => { if (r > bestR) { best = k; bestR = r } })
          if (best) setActive(best)
        },
        { threshold: [0, 0.25, 0.5], rootMargin: '-80px 0px -60% 0px' },
      )
      obs.observe(el)
      observers.push(obs)
    })
    return () => observers.forEach((o) => o.disconnect())
  }, [ids])

  return active
}

// ── CodeBlock ────────────────────────────────────────────────────────

function CodeBlock({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = useCallback(async () => {
    try { await navigator.clipboard.writeText(code); setCopied(true); setTimeout(() => setCopied(false), 2000) } catch {}
  }, [code])

  return (
    <div className="group/code relative">
      <pre className="overflow-x-auto rounded-md border border-border bg-muted/50 p-4 text-[13px] leading-relaxed">
        <code>{code}</code>
      </pre>
      <Button
        variant="ghost" size="icon"
        className="absolute right-2 top-2 h-6 w-6 text-muted-foreground opacity-0 transition-opacity hover:text-foreground group-hover/code:opacity-100"
        onClick={handleCopy}
      >
        {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
      </Button>
    </div>
  )
}

// ── LangTabs ─────────────────────────────────────────────────────────

function LangTabs({ exampleKey }: { exampleKey: string }) {
  const [lang, setLang] = useState<Lang>('curl')
  const example = EXAMPLES[exampleKey]
  if (!example) return null

  return (
    <div className="space-y-2">
      <div className="inline-flex gap-0.5 rounded-md border bg-muted/30 p-0.5">
        {LANGS.map((l) => (
          <button key={l} onClick={() => setLang(l)}
            className={cn(
              'rounded-[5px] px-2.5 py-0.5 text-xs transition-all',
              lang === l ? 'bg-background text-foreground shadow-xs' : 'text-muted-foreground hover:text-foreground',
            )}
          >{LANG_LABELS[l]}</button>
        ))}
      </div>
      <CodeBlock code={example[lang]} />
    </div>
  )
}

// ── ParamTable ───────────────────────────────────────────────────────

function ParamTable({ title, params }: { title: string; params: Param[] }) {
  if (!params.length) return null
  return (
    <div className="space-y-1.5">
      <p className="text-xs font-medium text-muted-foreground">{title}</p>
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50 hover:bg-muted/50">
              <TableHead className="h-8 text-xs">参数</TableHead>
              <TableHead className="h-8 text-xs">类型</TableHead>
              <TableHead className="h-8 text-xs">必填</TableHead>
              <TableHead className="h-8 text-xs">说明</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {params.map((p) => (
              <TableRow key={p.name} className="hover:bg-transparent">
                <TableCell className="py-1.5"><code className="text-xs">{p.name}</code></TableCell>
                <TableCell className="py-1.5"><span className="text-xs text-muted-foreground">{p.type}</span></TableCell>
                <TableCell className="py-1.5">
                  {p.required
                    ? <span className="text-xs text-rose-500">是</span>
                    : <span className="text-xs text-muted-foreground">否</span>}
                </TableCell>
                <TableCell className="py-1.5 text-xs text-muted-foreground">{p.desc}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}

// ── EndpointCard ─────────────────────────────────────────────────────

function EndpointCard({ ep }: { ep: Endpoint }) {
  const hasParams = (ep.requestParams?.length || 0) > 0 || (ep.requestBody?.length || 0) > 0 || (ep.response?.length || 0) > 0

  return (
    <div className="space-y-3 rounded-lg border p-4">
      <div className="flex flex-wrap items-center gap-2">
        <Badge variant="outline" className={cn('font-mono text-[10px] font-bold tracking-wider', METHOD_STYLES[ep.method])}>
          {ep.method}
        </Badge>
        <code className="font-mono text-xs">{ep.path}</code>
        <span className="text-sm text-muted-foreground">{ep.desc}</span>
      </div>

      {hasParams && (
        <div className="space-y-3">
          {ep.requestParams && <ParamTable title="Query 参数" params={ep.requestParams} />}
          {ep.requestBody && <ParamTable title="请求体 (JSON)" params={ep.requestBody} />}
          {ep.response && ep.response.length > 0 && <ParamTable title="响应字段" params={ep.response} />}
          {ep.response && ep.response.length === 0 && (
            <p className="text-xs text-muted-foreground">响应：无额外数据，仅返回操作状态。</p>
          )}
        </div>
      )}
    </div>
  )
}

// ── EndpointSection ──────────────────────────────────────────────────

const EXAMPLE_MAP: Record<string, string> = {
  keys: 'createKey',
  status: 'getKeyStatus',
  consume: 'consume',
  adjust: 'adjustBalance',
  export: 'consumeResult',
}

function EndpointSection({ groupKey }: { groupKey: string }) {
  const [open, setOpen] = useState(true)
  const group = GROUPS[groupKey]
  if (!group) return null
  const exampleKey = EXAMPLE_MAP[groupKey]

  return (
    <section id={groupKey} className="scroll-mt-4">
      <Collapsible open={open} onOpenChange={setOpen}>
        <div className="mb-1 flex items-center justify-between">
          <h3 className="text-lg font-semibold">{group.title}</h3>
          <CollapsibleTrigger asChild>
            <Button variant="ghost" size="sm" className="h-7 gap-1 text-xs text-muted-foreground">
              {open ? '收起' : '展开'}
              <ChevronDown className={cn('h-3 w-3 transition-transform', !open && '-rotate-90')} />
            </Button>
          </CollapsibleTrigger>
        </div>
        <p className="mb-4 text-sm text-muted-foreground">{group.description}</p>

        <CollapsibleContent>
          <div className="space-y-3">
            {group.endpoints.map((ep) => (
              <EndpointCard key={ep.method + ep.path} ep={ep} />
            ))}
          </div>

          {exampleKey && EXAMPLES[exampleKey] && (
            <div className="mt-4">
              <p className="mb-2 text-xs font-medium text-muted-foreground">示例代码</p>
              <LangTabs exampleKey={exampleKey} />
            </div>
          )}
        </CollapsibleContent>
      </Collapsible>
      <Separator className="mt-6" />
    </section>
  )
}

// ── Main ─────────────────────────────────────────────────────────────

const ALL_IDS = SIDEBAR_GROUPS.flatMap((g) => g.items.map((i) => i.id))

export default function ServiceApiDocs() {
  const contentRef = useRef<HTMLDivElement>(null)
  const active = useActiveSection(ALL_IDS)

  const scrollTo = (id: string) => {
    contentRef.current?.querySelector(`#${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  return (
    <Card className="flex min-h-[600px] overflow-hidden">
      {/* Sidebar */}
      <aside className="hidden w-48 shrink-0 border-r md:block">
        <ScrollArea className="h-full">
          <div className="space-y-5 p-4">
            {SIDEBAR_GROUPS.map((group) => (
              <div key={group.label}>
                <p className="mb-1.5 px-2 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">
                  {group.label}
                </p>
                <nav className="space-y-px">
                  {group.items.map(({ id, label }) => (
                    <button key={id} onClick={() => scrollTo(id)}
                      className={cn(
                        'flex w-full items-center rounded-md px-2.5 py-1.5 text-sm transition-colors',
                        active === id
                          ? 'bg-muted font-medium text-foreground'
                          : 'text-muted-foreground hover:text-foreground',
                      )}
                    >{label}</button>
                  ))}
                </nav>
              </div>
            ))}
          </div>
        </ScrollArea>
      </aside>

      {/* Content */}
      <div ref={contentRef} className="flex-1 overflow-auto">
        <div className="space-y-8 p-6">
          {/* 快速开始 */}
          <section id="quickstart" className="scroll-mt-4">
            <h3 className="mb-4 text-lg font-semibold">快速开始</h3>
            <div className="mb-4 space-y-2 text-sm text-muted-foreground">
              <p>服务账号通过 <code className="rounded bg-muted px-1 py-0.5 text-xs">X-Service-Key</code> 请求头调用 <code className="rounded bg-muted px-1 py-0.5 text-xs">/service/*</code> 接口。所有操作自动限定在当前租户内。</p>
              <p><span className="font-medium text-foreground">Base URL：</span><code className="rounded bg-muted px-1 py-0.5 text-xs">https://your-domain.com/api</code></p>
              <p>在上方「服务账号管理」创建账号后保存密钥（<code className="text-xs">svc_</code> 开头），后续请求均需携带。</p>
            </div>
            <p className="mb-2 text-sm font-medium">最小示例</p>
            <LangTabs exampleKey="createKey" />
          </section>

          <Separator />

          {/* 认证方式 */}
          <section id="auth" className="scroll-mt-4">
            <h3 className="mb-3 text-lg font-semibold">认证方式</h3>
            <p className="mb-3 text-sm text-muted-foreground">
              所有 <code className="rounded bg-muted px-1 py-0.5 text-xs">/service/*</code> 请求均需携带 Header：
            </p>
            <CodeBlock code="X-Service-Key: sk_your_service_key_here" />
            <p className="mt-4 mb-2 text-sm font-medium">认证失败响应</p>
            <CodeBlock code={'{"code": 3001, "message": "服务账号密钥无效", "data": null}'} />
          </section>

          <Separator />

          {/* 端点分组 */}
          {Object.keys(GROUPS).map((key) => (
            <EndpointSection key={key} groupKey={key} />
          ))}

          {/* 幂等机制 */}
          <section id="idempotency" className="scroll-mt-4">
            <h3 className="mb-3 text-lg font-semibold">幂等机制</h3>
            <div className="mb-4 space-y-2 text-sm text-muted-foreground">
              <p>消费和额度调整接口支持幂等调用，防止因网络重试导致重复操作。</p>
              <p>请求携带 <code className="rounded bg-muted px-1 py-0.5 text-xs">X-Request-Id</code>（UUID），服务端在 Redis 中缓存结果 24 小时。相同 ID 的后续请求直接返回缓存结果。</p>
              <p>通过 <code className="rounded bg-muted px-1 py-0.5 text-xs">GET /service/consume-result?request_id=xxx</code> 查询操作结果，优先查 Redis，过期后 fallback 到数据库日志。</p>
            </div>
            <LangTabs exampleKey="consumeResult" />
          </section>

          <Separator />

          {/* 错误码 */}
          <section id="errors" className="scroll-mt-4">
            <h3 className="mb-3 text-lg font-semibold">错误码</h3>
            <div className="overflow-hidden rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50 hover:bg-muted/50">
                    <TableHead className="h-8 text-xs">Code</TableHead>
                    <TableHead className="h-8 text-xs">说明</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {[
                    ['0', '成功'],
                    ['3001', '服务账号密钥无效'],
                    ['3002', '服务账号已被禁用'],
                    ['4004', '租户不存在或已过期'],
                    ['1001', '卡密不存在'],
                    ['1002', '卡密额度不足'],
                    ['1003', '卡密已被禁用'],
                    ['1004', '无效的扣减数量'],
                    ['1005', '无效的调整量（不能为 0）'],
                    ['9999', '系统内部错误'],
                  ].map(([code, msg]) => (
                    <TableRow key={code} className="hover:bg-transparent">
                      <TableCell className="py-1.5"><code className="text-xs">{code}</code></TableCell>
                      <TableCell className="py-1.5 text-xs text-muted-foreground">{msg}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </section>
        </div>
      </div>
    </Card>
  )
}
