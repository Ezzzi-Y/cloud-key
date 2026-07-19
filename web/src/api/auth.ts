import api from './client'
import type { ApiResponse, LoginStep1Data, LoginResponse, UserProfile } from '@/types'

export function login(username: string, password: string) {
  return api.post<unknown, ApiResponse<LoginStep1Data>>('/auth/login', { username, password })
}

export function verify2FA(userId: number, code: string, preAuthToken: string) {
  return api.post<unknown, ApiResponse<LoginResponse>>('/auth/verify-2fa', { user_id: userId, code, pre_auth_token: preAuthToken })
}

export function setupTOTPInit(userId: number) {
  return api.post<unknown, ApiResponse<{ secret: string; url: string }>>('/auth/totp/setup-init', { user_id: userId })
}

export function confirmTOTPInit(userId: number, code: string, preAuthToken: string) {
  return api.post<unknown, ApiResponse<LoginResponse>>('/auth/totp/confirm-init', { user_id: userId, code, pre_auth_token: preAuthToken })
}

export function setupTOTP(role: string) {
  return api.post<unknown, ApiResponse<{ secret: string; url: string }>>(`/${role}/totp/setup`)
}

export function confirmTOTP(role: string, code: string) {
  return api.post<unknown, ApiResponse<null>>(`/${role}/totp/confirm`, { code })
}

export function getProfile(role: string) {
  return api.get<unknown, ApiResponse<UserProfile>>(`/${role}/profile`)
}

export function changePassword(role: string, oldPassword: string, newPassword: string) {
  return api.put<unknown, ApiResponse<null>>(`/${role}/password`, {
    old_password: oldPassword,
    new_password: newPassword,
  })
}
