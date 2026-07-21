import api from './client'
import type { ApiResponse, PaginatedData, UsageLog, UsageLogQuery, LoginLog, BalanceLog, BalanceLogQuery } from '@/types'

export function listUsageLogs(query: UsageLogQuery) {
  return api.get<unknown, ApiResponse<PaginatedData<UsageLog>>>('/tenant/usage-logs', { params: query })
}

export function exportUsageLogs(query: Omit<UsageLogQuery, 'page' | 'page_size'>) {
  return api.get<unknown, ApiResponse<UsageLog[]>>('/tenant/usage-logs/export', { params: query })
}

export function listLoginLogs(role: string, page = 1, pageSize = 20) {
  return api.get<unknown, ApiResponse<PaginatedData<LoginLog>>>(`/${role}/login-logs`, {
    params: { page, page_size: pageSize },
  })
}

export function listBalanceLogs(query: BalanceLogQuery) {
  return api.get<unknown, ApiResponse<PaginatedData<BalanceLog>>>('/tenant/balance-logs', { params: query })
}

export function exportBalanceLogs(query: Omit<BalanceLogQuery, 'page' | 'page_size'>) {
  return api.get<unknown, ApiResponse<BalanceLog[]>>('/tenant/balance-logs/export', { params: query })
}
