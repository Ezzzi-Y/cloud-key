import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('ck_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  // 每个请求自动生成 request_id（用于幂等 + 追踪）
  config.headers['X-Request-Id'] = crypto.randomUUID()
  return config
})

api.interceptors.response.use(
  (res) => res.data,
  (err) => {
    if (err.response?.status === 401 && window.location.pathname !== '/login') {
      localStorage.removeItem('ck_token')
      window.location.href = '/login'
    }
    return Promise.reject(err.response?.data || err)
  },
)

export default api
