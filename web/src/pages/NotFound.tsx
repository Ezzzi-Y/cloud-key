import { Button } from '@/components/ui/button'
import { useNavigate } from 'react-router-dom'

export default function NotFound() {
  const navigate = useNavigate()
  return (
    <div className="flex h-screen flex-col items-center justify-center gap-4">
      <h1 className="text-7xl font-bold text-muted-foreground">404</h1>
      <p className="text-lg text-muted-foreground">页面不存在</p>
      <Button onClick={() => navigate('/')}>返回首页</Button>
    </div>
  )
}
