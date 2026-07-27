export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export type UserRole = 'super_admin' | 'tenant_admin'
export type TenantStatus = 'active' | 'expired' | 'disabled'

export interface LoginResponse {
  token: string
  token_type: string
  role: UserRole
  tenant_id: number | null
  username: string
  tenant_status?: TenantStatus | null
  tenant_expire_at?: string | null
}

export interface LoginStep1Data {
  require_totp: boolean
  need_setup?: boolean
  user_id: number
  pre_auth_token: string
}

export interface Tenant {
  id: number
  name: string
  status: 'active' | 'expired' | 'disabled'
  expire_at: string | null
  key_prefix: string
  key_length: number
  key_suffix_length: number
  created_at: string
  updated_at: string
}

export interface TenantListItem extends Tenant {
  key_count: number
  user_count: number
}

export type KeyStatus = 'active' | 'exhausted' | 'disabled' | 'expired'

export interface Key {
  id: number
  tenant_id: number
  alias: string
  key_prefix: string
  key_suffix: string
  remaining_amount: number
  status: KeyStatus
  created_by: string
  created_at: string
  updated_at: string
  used_at: string | null
  expire_at: string | null
  max_usage: number | null
}

export interface CreateKeyRequest {
  alias: string
  remaining_amount: number
}

export interface UpdateKeyRequest {
  alias?: string
}

export interface KeyListParams {
  page?: number
  page_size?: number
  status?: KeyStatus
  alias?: string
  key_suffix?: string
}

export interface KeyStatusResult {
  alias: string
  remaining_amount: number
  status: KeyStatus
  created_at: string
  used_at: string | null
}

export interface ConsumeKeyRequest {
  key: string
  amount?: number
}

export interface ConsumeKeyResult {
  remaining_amount: number
  status: KeyStatus
  used_up: boolean
  exhausted?: boolean
}

export interface ServiceAccount {
  id: number
  tenant_id: number
  name: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface UsageLog {
  id: number
  tenant_id: number
  key_id: number
  key_alias: string
  key_suffix: string
  amount: number
  ip: string
  user_agent: string
  response_status: number
  created_at: string
}

export interface UsageLogQuery {
  page?: number
  page_size?: number
  alias?: string
  key_suffix?: string
  start_time?: string
  end_time?: string
}

export interface LoginLog {
  id: number
  user_id: number
  tenant_id: number | null
  ip: string
  user_agent: string
  status: 'success' | 'failed'
  created_at: string
}

export interface DashboardData {
  key_count: number
  today_calls: number
  week_calls: number
  month_calls: number
  recent_logs: UsageLog[]
  key_status_breakdown: Record<string, number>
  can_refresh: boolean
  next_refresh_at: string | null
}

export interface KeyOverview {
  total_keys: number
  keys_by_status: Record<string, number>
  total_remaining_amount: number
}

export interface TrendPoint {
  date: string
  calls: number
  amount: number
}

export interface TrendData {
  points: TrendPoint[]
}

export interface TopItem {
  key_alias?: string
  key_suffix?: string
  ip?: string
  count: number
  total_amount: number
}

export interface KeyUsagePoint {
  date: string
  calls: number
  amount: number
}

export interface KeyUsageStats {
  key_alias: string
  key_suffix: string
  points: KeyUsagePoint[]
  can_refresh: boolean
  next_refresh_at: string | null
}

export interface BalanceLog {
  id: number
  tenant_id: number
  key_id: number
  key_alias: string
  key_suffix: string
  delta: number
  before_amount: number
  after_amount: number
  operator: string
  remark: string
  created_at: string
}

export interface BalanceLogQuery {
  page?: number
  page_size?: number
  alias?: string
  key_suffix?: string
  start_time?: string
  end_time?: string
}

export interface AdjustBalanceRequest {
  delta: number
  remark?: string
}

export interface AdjustBalanceResult {
  before_amount: number
  after_amount: number
}

export interface ConsumeResultQuery {
  source: string
  request_id: string
  key_id?: number
  key_alias?: string
  key_suffix?: string
  amount?: number
  delta?: number
  before_amount?: number
  after_amount?: number
  ip?: string
  operator?: string
  remark?: string
  created_at?: string
  http_status?: number
  code?: number
  data?: ConsumeKeyResult
}

export interface SysConfig {
  id: number
  key: string
  value: string
  description: string
  updated_at: string
}

export interface UserProfile {
  id: number
  username: string
  role: UserRole
  tenant_id: number | null
  totp_setup: boolean
  is_active: boolean
  created_at: string
  tenant_status?: TenantStatus
  tenant_expire_at?: string | null
}
