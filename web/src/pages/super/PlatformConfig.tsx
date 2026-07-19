import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getConfigs, updateConfigs, type UpdateConfigItem } from '@/api/config'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { toast } from 'sonner'

interface Config {
  id: number
  key: string
  value: string
  description: string
}

export default function PlatformConfig() {
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState<Record<string, string>>({})
  const { data: configs, isLoading } = useQuery({
    queryKey: ['configs'],
    queryFn: () => getConfigs().then((r) => (r.code === 0 ? r.data : [])),
  })
  useEffect(() => {
    if (configs) { const m: Record<string, string> = {}; (configs as Config[]).forEach((c) => { m[c.key] = c.value }); setEditing(m) }
  }, [configs])

  const mutation = useMutation({
    mutationFn: (items: UpdateConfigItem[]) => updateConfigs(items),
    onSuccess: (res) => {
      if (res.code === 0) { toast.success('配置已保存'); queryClient.invalidateQueries({ queryKey: ['configs'] }) }
      else toast.error(res.message)
    },
  })

  const handleSave = () => {
    if (!configs) return
    const items: UpdateConfigItem[] = (configs as Config[]).map((c) => ({ key: c.key, value: editing[c.key] || '', description: c.description }))
    mutation.mutate(items)
  }

  if (isLoading) return <div className="flex items-center justify-center py-12">加载中...</div>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">平台配置</h2>
        <Button onClick={handleSave} disabled={mutation.isPending}>{mutation.isPending ? '保存中...' : '保存全部'}</Button>
      </div>
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow><TableHead>配置键</TableHead><TableHead>值</TableHead><TableHead>说明</TableHead></TableRow>
            </TableHeader>
            <TableBody>
              {(configs as Config[])?.map((c) => (
                <TableRow key={c.id}>
                  <TableCell className="font-mono text-sm">{c.key}</TableCell>
                  <TableCell><Input value={editing[c.key] || ''} onChange={(e) => setEditing((prev) => ({ ...prev, [c.key]: e.target.value }))} className="max-w-xs" /></TableCell>
                  <TableCell className="text-muted-foreground">{c.description}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
