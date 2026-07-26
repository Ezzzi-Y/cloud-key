import api from './client'
import type { ApiResponse, DashboardData, KeyOverview, TrendPoint, TopItem } from '@/types'

export function getDashboard() {
  return api.get<unknown, ApiResponse<DashboardData>>('/tenant/stats/dashboard')
}

export function getOverview(startDate?: string, endDate?: string) {
  return api.get<unknown, ApiResponse<KeyOverview>>('/tenant/stats/overview', {
    params: startDate && endDate ? { start_date: startDate, end_date: endDate } : {},
  })
}

export function getTrends(period: 'today' | 'week' | 'month' = 'today') {
  return api.get<unknown, ApiResponse<TrendPoint[]>>('/tenant/stats/trends', { params: { period } })
}

export function getTopKeys() {
  return api.get<unknown, ApiResponse<TopItem[]>>('/tenant/stats/top-keys')
}

export function getTopAmount() {
  return api.get<unknown, ApiResponse<TopItem[]>>('/tenant/stats/top-amount')
}

export function refreshTopStats() {
  return api.post<unknown, ApiResponse<null>>('/tenant/stats/refresh-top')
}
