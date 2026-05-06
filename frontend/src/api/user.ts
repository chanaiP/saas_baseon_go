import http, { unwrap } from './http'
import type { ApiResponse } from './types'
import type { Paginated } from './types'

export interface UserRow {
  id: number
  tenant_id: number
  employee_no: string
  phone: string | null
  name: string
  email: string | null
  avatar_url: string | null
  status: number
  company_id: number | null
  department_id: number | null
  department_ids?: number[]
  position_ids?: number[]
  role_ids: number[]
  is_platform_admin?: boolean
  created_at?: string
}

export interface AssignableRoleRow {
  id: number
  code: string
  name: string
}

/** 用户管理表单下拉：仅需用户管理菜单权限，不依赖角色管理菜单 */
export async function fetchAssignableRoles(params?: { skip?: number; limit?: number; kw?: string }) {
  return unwrap(http.get<ApiResponse<Paginated<AssignableRoleRow>>>('/api/users/assignable-roles', { params }))
}

export async function fetchUsers(params: {
  skip?: number
  limit?: number
  keyword?: string
  status?: number
  company_id?: number
  department_id?: number
}) {
  return unwrap(http.get<ApiResponse<Paginated<UserRow>>>('/api/users', { params }))
}

export async function createUser(body: {
  employee_no: string
  password: string
  name: string
  phone?: string
  email?: string
  company_id?: number | null
  department_id?: number | null
  department_ids?: number[]
  position_ids?: number[]
  role_ids?: number[]
  status?: number
}) {
  return unwrap(http.post<ApiResponse<UserRow>>('/api/users', body))
}

export async function updateUser(
  id: number,
  body: Partial<{
    phone: string | null
    name: string
    email: string | null
    company_id: number | null
    department_id: number | null
    department_ids: number[]
    position_ids: number[]
    role_ids: number[]
    status: number
    is_platform_admin?: boolean
  }>,
) {
  return unwrap(http.put<ApiResponse<UserRow>>(`/api/users/${id}`, body))
}

export async function resetUserPassword(id: number) {
  return unwrap(http.put<ApiResponse<{ new_password: string }>>(`/api/users/${id}/password`))
}

export async function deleteUser(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/users/${id}`))
}
