import api from './client'
import type { ApiResponse, PaginatedData, Key, CreateKeyRequest, UpdateKeyRequest, KeyListParams, KeyStatusResult, ConsumeKeyRequest, ConsumeKeyResult } from '@/types'

export function getKeyStatus(sk: string) {
  return api.get<unknown, ApiResponse<KeyStatusResult>>('/key/status', { params: { sk } })
}

export function consumeKey(data: ConsumeKeyRequest) {
  return api.post<unknown, ApiResponse<ConsumeKeyResult>>('/key/consume', data)
}

export function listKeys(params: KeyListParams) {
  return api.get<unknown, ApiResponse<PaginatedData<Key>>>('/tenant/keys', { params })
}

export function createKey(data: CreateKeyRequest) {
  return api.post<unknown, ApiResponse<{ raw_key: string; key: Key }>>('/tenant/keys', data)
}

export function getKeyDetail(id: number) {
  return api.get<unknown, ApiResponse<Key>>(`/tenant/keys/${id}`)
}

export function updateKey(id: number, data: UpdateKeyRequest) {
  return api.patch<unknown, ApiResponse<null>>(`/tenant/keys/${id}`, data)
}

export function disableKey(id: number) {
  return api.patch<unknown, ApiResponse<null>>(`/tenant/keys/${id}/disable`)
}

export function enableKey(id: number) {
  return api.patch<unknown, ApiResponse<null>>(`/tenant/keys/${id}/enable`)
}

export function deleteKey(id: number) {
  return api.delete<unknown, ApiResponse<null>>(`/tenant/keys/${id}`)
}

export function exportKeysCSV() {
  return api.get('/tenant/keys/export', { responseType: 'blob' })
}

export function exportKeysJSON() {
  return api.get<unknown, ApiResponse<Key[]>>('/tenant/keys/export/json')
}
