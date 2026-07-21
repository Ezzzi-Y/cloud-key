import api from './client'
import type { ApiResponse, ServiceAccount } from '@/types'

export interface CreateServiceAccountResult {
  id: number
  name: string
  raw_key: string
  is_active: boolean
  created_at: string
}

export function listServiceAccounts() {
  return api.get<unknown, ApiResponse<ServiceAccount[]>>('/tenant/service-accounts')
}

export function createServiceAccount(name: string) {
  return api.post<unknown, ApiResponse<CreateServiceAccountResult>>('/tenant/service-accounts', { name })
}

export function toggleServiceAccount(id: number, isActive: boolean) {
  return api.patch<unknown, ApiResponse<null>>(`/tenant/service-accounts/${id}/toggle`, { is_active: isActive })
}

export function deleteServiceAccount(id: number) {
  return api.delete<unknown, ApiResponse<null>>(`/tenant/service-accounts/${id}`)
}
