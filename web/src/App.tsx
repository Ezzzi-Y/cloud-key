import { Routes, Route } from 'react-router-dom'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<div className="p-8 text-xl">CloudKey 管理后台加载中...</div>} />
    </Routes>
  )
}
