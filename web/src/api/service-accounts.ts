import api from './client'
import type { ApiResponse, ServiceAccount } from '@/types'

export interface CreateServiceAccountResult {
  account: ServiceAccount
  raw_key: string
}

export function listServiceAccounts() {
  return api.get<unknown, ApiResponse<ServiceAccount[]>>('/tenant/service-accounts')
}

export function createServiceAccount(name: string) {
  return api.post<unknown, ApiResponse<CreateServiceAccountResult>>('/tenant/service-accounts', { name })
}

export function toggleServiceAccount(id: number) {
  return api.patch<unknown, ApiResponse<null>>(`/tenant/service-accounts/${id}/toggle`)
}

export function deleteServiceAccount(id: number) {
  return api.delete<unknown, ApiResponse<null>>(`/tenant/service-accounts/${id}`)
}
