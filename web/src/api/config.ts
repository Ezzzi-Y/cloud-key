import api from './client'
import type { ApiResponse, SysConfig } from '@/types'

export interface UpdateConfigItem {
  key: string
  value: string
  description: string
}

export function getConfigs() {
  return api.get<unknown, ApiResponse<SysConfig[]>>('/super/configs')
}

export function updateConfigs(items: UpdateConfigItem[]) {
  return api.put<unknown, ApiResponse<null>>('/super/configs', items)
}
