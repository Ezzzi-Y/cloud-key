import api from './client'
import type { ApiResponse, TenantListItem, Tenant } from '@/types'

export interface CreateTenantResult {
  tenant: Tenant
  admin_username: string
  admin_password: string
}

export interface UpdateTenantRequest {
  name?: string
  status?: 'active' | 'expired' | 'disabled'
  expire_at?: string
}

export function listTenants() {
  return api.get<unknown, ApiResponse<TenantListItem[]>>('/super/tenants')
}

export function createTenant(name: string) {
  return api.post<unknown, ApiResponse<CreateTenantResult>>('/super/tenants', { name })
}

export function getTenant(id: number) {
  return api.get<unknown, ApiResponse<TenantListItem>>(`/super/tenants/${id}`)
}

export function updateTenant(id: number, data: UpdateTenantRequest) {
  return api.patch<unknown, ApiResponse<null>>(`/super/tenants/${id}`, data)
}

export function resetTenantPassword(id: number) {
  return api.patch<unknown, ApiResponse<{ new_password: string }>>(`/super/tenants/${id}/reset-password`)
}
