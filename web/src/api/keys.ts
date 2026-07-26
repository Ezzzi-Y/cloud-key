import api from './client'
import type { ApiResponse, PaginatedData, Key, CreateKeyRequest, UpdateKeyRequest, KeyListParams, KeyStatusResult, ConsumeKeyRequest, ConsumeKeyResult, AdjustBalanceRequest, AdjustBalanceResult } from '@/types'

export function getKeyStatus(sk: string) {
  return api.get<unknown, ApiResponse<KeyStatusResult>>('/tenant/keys/status', { params: { sk } })
}

export function consumeKey(data: ConsumeKeyRequest) {
  return api.post<unknown, ApiResponse<ConsumeKeyResult>>('/tenant/keys/consume', data)
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
  return api.get('/tenant/keys/export', { responseType: 'text' })
}

export function adjustKeyBalance(id: number, data: AdjustBalanceRequest) {
  return api.post<unknown, ApiResponse<AdjustBalanceResult>>(`/tenant/keys/${id}/adjust-balance`, data)
}

export function exportKeysJSON() {
  return api.get<unknown, ApiResponse<Key[]>>('/tenant/keys/export/json')
}

export interface KeyConfig {
  key_prefix: string
  key_length: number
  key_suffix_length: number
}

export interface UpdateKeyConfigRequest {
  key_prefix?: string
  key_length?: number
  key_suffix_length?: number
}

export function getKeyConfig() {
  return api.get<unknown, ApiResponse<KeyConfig>>('/tenant/key-config')
}

export function updateKeyConfig(data: UpdateKeyConfigRequest) {
  return api.patch<unknown, ApiResponse<KeyConfig>>('/tenant/key-config', data)
}
