import api from './client'
import type { ApiResponse, DashboardData, KeyOverview, TrendData, TopItem } from '@/types'

export function getDashboard() {
  return api.get<unknown, ApiResponse<DashboardData>>('/tenant/stats/dashboard')
}

export function getOverview(startDate?: string, endDate?: string) {
  return api.get<unknown, ApiResponse<KeyOverview>>('/tenant/stats/overview', {
    params: startDate && endDate ? { start_date: startDate, end_date: endDate } : {},
  })
}

export function getTrends(period: 'today' | 'week' | 'month' = 'today') {
  return api.get<unknown, ApiResponse<TrendData>>('/tenant/stats/trends', { params: { period } })
}

export function getTopKeys() {
  return api.get<unknown, ApiResponse<TopItem[]>>('/tenant/stats/top-keys')
}

export function getTopIPs() {
  return api.get<unknown, ApiResponse<TopItem[]>>('/tenant/stats/top-ips')
}
