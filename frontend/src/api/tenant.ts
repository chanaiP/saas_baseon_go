import http, { unwrap } from './http'
import type { ApiResponse } from './types'
import type { Paginated } from './types'
import type { TenantQuotaOverride, TenantSubscription, TenantSubscriptionPayload } from './plan'

export interface Tenant {
  id: number
  code: string
  name: string
  status: number
  start_date?: string | null
  expire_date?: string | null
  max_companies?: number
  max_stores?: number
  max_business_units?: number
  max_users?: number
  used_companies?: number
  used_stores?: number
  used_business_units?: number
  used_users?: number
  /** 当前订阅套餐名称（列表/详情由后端填充） */
  plan_name?: string | null
  /** 当前订阅套餐编码（用于前端区分胶囊配色） */
  plan_code?: string | null
  contact_name?: string | null
  contact_phone?: string | null
  company_count?: number
  created_at?: string
  brand_display_name?: string | null
  brand_logo_data?: string | null
}

export interface CompanyTreeNode {
  id: number
  code: string | null
  name: string
  company_type: string | null
  status: number
  children: CompanyTreeNode[]
}

export interface TenantOrgQuotaRecord {
  id: number
  code?: string | null
  name: string
  node_type: string
  company_name?: string | null
  status: number
}

export interface TenantBusinessUnitQuotaRecord {
  id: number
  code: string
  name: string
  bu_type?: string | null
  status: number
}

export async function fetchTenants(skip = 0, limit = 20) {
  return unwrap(http.get<ApiResponse<Paginated<Tenant>>>('/api/tenants', { params: { skip, limit } }))
}

/** 单条详情（含从主管理员回填的联系人，供编辑表单使用） */
export async function fetchTenant(id: number) {
  return unwrap(http.get<ApiResponse<Tenant>>(`/api/tenants/${id}`))
}

export interface TenantCreatePayload {
  code: string
  name: string
  status?: number
  admin_name: string
  admin_employee_no: string
  admin_phone?: string
  admin_password: string
}

export async function createTenant(body: TenantCreatePayload) {
  return unwrap(http.post<ApiResponse<Tenant>>('/api/tenants', body))
}

export interface TenantCreateWithPackagePayload {
  tenant: TenantCreatePayload
  package: TenantPackageConfigPayload
}

/** 向导合并提交：创建主体与套餐在同一请求内完成（中途关闭不会落库半成品主体） */
export async function createTenantWithPackage(body: TenantCreateWithPackagePayload) {
  return unwrap(http.post<ApiResponse<Tenant>>('/api/tenants/with-package', body))
}

export async function updateTenant(
  id: number,
  body: Partial<Pick<Tenant, 'name' | 'status'>> & {
    contact_name?: string
    contact_phone?: string
    brand_display_name?: string | null
    brand_logo_data?: string | null
  },
) {
  return unwrap(http.put<ApiResponse<Tenant>>(`/api/tenants/${id}`, body))
}

/** 列表快捷启停：需 tenant:status（与完整编辑 tenant:edit 拆分） */
export async function patchTenantStatus(id: number, status: 0 | 1) {
  return unwrap(http.patch<ApiResponse<Tenant>>(`/api/tenants/${id}/status`, { status }))
}

export interface TenantPackageQuotaPayload {
  quota_id: number
  quota_value: number
}

export interface TenantPackageConfigPayload extends TenantSubscriptionPayload {
  quotas: TenantPackageQuotaPayload[]
}

export interface TenantPackageConfigResult {
  tenant_id: number
  subscription: TenantSubscription
  quotas: { tenant_id: number; overrides: TenantQuotaOverride[] }
}

export async function saveTenantPackageConfig(id: number, body: TenantPackageConfigPayload) {
  return unwrap(http.put<ApiResponse<TenantPackageConfigResult>>(`/api/tenants/${id}/package-config`, body))
}

export async function deleteTenant(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/tenants/${id}`))
}

export async function fetchTenantCompanies(tenantId: number) {
  return unwrap(
    http.get<ApiResponse<{ tenant_id: number; billing_unit_count: number; companies: CompanyTreeNode[] }>>(
      `/api/tenants/${tenantId}/companies`,
    ),
  )
}

export async function fetchTenantQuotaRecords(tenantId: number) {
  return unwrap(
    http.get<ApiResponse<{
      tenant_id: number
      companies: TenantOrgQuotaRecord[]
      stores: TenantOrgQuotaRecord[]
      business_units: TenantBusinessUnitQuotaRecord[]
    }>>(`/api/tenants/${tenantId}/quota-records`),
  )
}

export interface TenantPrimaryAdminBrief {
  employee_no: string
  phone?: string | null
  name: string
}

export interface TenantPrimaryAdminPasswordResetResult extends TenantPrimaryAdminBrief {
  new_password: string
}

/** 系统管理员：查询主体主管理员账号（工号/手机），用于重置前展示 */
export async function fetchTenantPrimaryAdmin(tenantId: number) {
  return unwrap(http.get<ApiResponse<TenantPrimaryAdminBrief>>(`/api/tenants/${tenantId}/primary-admin`))
}

/** 系统管理员：随机重置该主体主管理员登录密码，响应一次性返回明文 */
export async function resetTenantPrimaryAdminPassword(tenantId: number) {
  return unwrap(
    http.put<ApiResponse<TenantPrimaryAdminPasswordResetResult>>(
      `/api/tenants/${tenantId}/primary-admin/password`,
    ),
  )
}
