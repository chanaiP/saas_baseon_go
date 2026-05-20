import http, { unwrap } from './http'
import type { ApiResponse } from './types'

export interface LoginPayload {
  account: string
  password: string
  captcha_id?: string
  captcha_code?: string
  /** 可选：缩小登录空间范围 */
  tenant_code?: string
  /** 多空间歧义时二次提交 */
  tenant_id?: number
}

export interface RegisterPayload {
  account?: string
  phone?: string
  password: string
  name?: string
  email?: string
}

export interface LoginTenantOption {
  tenant_id: number
  code: string
  name: string
}

export interface LoginData {
  token: string | null
  token_type: string
  captcha_required: boolean
  tenants?: LoginTenantOption[]
}

export interface CaptchaData {
  captcha_id: string
  image_base64: string
}

export interface Profile {
  id: number
  account_id?: number | null
  identity_user_id?: number | null
  tenant_id: number
  tenant_type?: 'platform' | 'enterprise' | 'personal'
  member_type?: string
  employee_no: string
  phone: string | null
  name: string
  email: string | null
  avatar_url: string | null
  status: number
  company_id: number | null
  department_id: number | null
  role_ids: number[]
  /** 当前用户所属角色的业务编码，与租户内角色一致（如 admin） */
  role_codes?: string[]
  permission_codes: string[]
  is_platform_admin?: boolean
  is_tenant_admin?: boolean
  /** 所属主体为平台运营租户（可通过 RBAC 承担运维能力，不必具备超级管理员标记） */
  tenant_is_platform?: boolean
  shortcut_ids?: string[]
  subscription?: {
    plan_code: string
    plan_name: string
    status: string
    end_time?: string | null
  } | null
  features?: string[]
  quotas?: Record<string, number>
}

export async function fetchCaptcha() {
  return unwrap(http.get<ApiResponse<CaptchaData>>('/api/auth/captcha'))
}

/** 手机号登录前置：该手机号在哪些空间有账号（非手机号输入时后端返回空数组） */
export async function fetchPhoneLoginTenants(account: string) {
  return unwrap(
    http.get<ApiResponse<LoginTenantOption[]>>('/api/auth/phone-login-tenants', {
      params: { account: account.trim() },
    }),
  )
}

export type LoginOutcome =
  | { kind: 'success'; token: string }
  | { kind: 'captcha'; message: string }
  | { kind: 'pickTenant'; message: string; tenants: LoginTenantOption[] }

export async function login(payload: LoginPayload): Promise<LoginOutcome> {
  const { data } = await http.post<ApiResponse<LoginData>>('/api/auth/login', payload)
  if (data.code === 0 && data.data?.token) {
    return { kind: 'success', token: data.data.token }
  }
  if (data.code === 1 && data.data?.captcha_required) {
    return { kind: 'captcha', message: data.message || '需要验证码' }
  }
  if (data.code === 2 && data.data?.tenants?.length) {
    return {
      kind: 'pickTenant',
      message: data.message || '请选择空间',
      tenants: data.data.tenants,
    }
  }
  throw new Error(data.message || '登录失败')
}

export async function register(payload: RegisterPayload): Promise<LoginOutcome> {
  const { data } = await http.post<ApiResponse<LoginData>>('/api/auth/register', payload)
  if (data.code === 0 && data.data?.token) {
    return { kind: 'success', token: data.data.token }
  }
  throw new Error(data.message || '注册失败')
}

export interface SwitchableTenantOption {
  tenant_id: number
  tenant_code: string
  tenant_name: string
  user_id: number
  employee_no: string
  user_display_name: string
}

export async function fetchSwitchableTenants() {
  return unwrap(http.get<ApiResponse<SwitchableTenantOption[]>>('/api/auth/switchable-tenants'))
}

export async function switchTenant(body: { tenant_id: number }) {
  const { data } = await http.post<ApiResponse<LoginData>>('/api/auth/switch-tenant', body)
  if (data.code !== 0 || !data.data?.token) {
    throw new Error(data.message || '切换失败')
  }
  return data.data
}

export async function logout() {
  return unwrap(http.post<ApiResponse<Record<string, never>>>('/api/auth/logout'))
}

export async function fetchProfile() {
  return unwrap(http.get<ApiResponse<Profile>>('/api/users/me'))
}

export async function updateProfile(body: Partial<Pick<Profile, 'name' | 'phone' | 'email' | 'avatar_url'>>) {
  return unwrap(http.put<ApiResponse<Profile>>('/api/users/me', body))
}

export interface PasswordChangeBody {
  old_password: string
  new_password: string
  new_password_confirm: string
  captcha_id: string
  captcha_code: string
}

export async function updateMyPassword(body: PasswordChangeBody) {
  return unwrap(http.put<ApiResponse<Record<string, never>>>('/api/users/me/password', body))
}

export interface ShortcutIdsPayload {
  shortcut_ids: string[]
}

export async function fetchShortcuts() {
  return unwrap(http.get<ApiResponse<ShortcutIdsPayload>>('/api/users/me/preferences'))
}

export async function saveShortcuts(ids: string[]) {
  return unwrap(http.put<ApiResponse<ShortcutIdsPayload>>('/api/users/me/preferences', { shortcut_ids: ids }))
}
