import http, { unwrap } from './http'
import type { ApiResponse } from './types'
import type { Paginated } from './types'

export interface RoleDataOverride {
  permission_id: number
  data_scope: string
  custom_company_ids: number[]
  custom_department_ids: number[]
  custom_user_ids: number[]
  custom_business_unit_ids?: number[]
  bu_data_access_mode?: 'CURRENT_ORG_BU' | 'SPECIFIED_BU' | null
}

export interface RoleRow {
  id: number
  code: string
  name: string
  description: string | null
  permission_ids: number[]
  data_overrides: RoleDataOverride[]
  is_system?: boolean
  is_locked?: boolean
  can_edit?: boolean
  can_delete?: boolean
  can_config_perm?: boolean
}

export async function fetchRoles(skip = 0, limit = 50, kw?: string) {
  const params: Record<string, any> = { skip, limit }
  if (kw) params.kw = kw
  return unwrap(http.get<ApiResponse<Paginated<RoleRow>>>('/api/roles', { params }))
}

export async function fetchRole(id: number) {
  return unwrap(http.get<ApiResponse<RoleRow>>(`/api/roles/${id}`))
}

export async function createRole(body: { code: string; name: string; description?: string }) {
  return unwrap(http.post<ApiResponse<RoleRow>>('/api/roles', body))
}

export async function updateRole(
  id: number,
  body: Partial<{
    name: string
    description: string
  }>,
) {
  return unwrap(http.put<ApiResponse<RoleRow>>(`/api/roles/${id}`, body))
}

export async function updateRolePermissions(
  id: number,
  body: {
    permission_ids: number[]
    data_overrides: RoleDataOverride[]
  },
) {
  return unwrap(http.put<ApiResponse<RoleRow>>(`/api/roles/${id}/permissions`, body))
}

export async function deleteRole(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/roles/${id}`))
}
