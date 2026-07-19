import api from './client'
import type { ApiResponse, PaginatedData, UsageLog, UsageLogQuery, LoginLog } from '@/types'

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
