import http, { unwrap } from './http'
import type { ApiResponse } from './types'

/** node_type 与字典 org_node_type 的 value 一致（如 company、warehouse） */
export interface OrgNode {
  id: number
  node_type: string
  name: string
  code: string | null
  company_type?: string | null
  company_id?: number | null
  store_id?: number | null
  parent_id?: number | null
  status: number
  children: OrgNode[]
}

export async function fetchOrgTree() {
  return unwrap(http.get<ApiResponse<OrgNode[]>>('/api/organizations/tree'))
}

export async function createOrgNode(tenantId: number, body: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/org-nodes', body, { params: { tenant_id: tenantId } }))
}

export async function updateOrgNode(tenantId: number, nodeId: number, body: Record<string, unknown>) {
  return unwrap(
    http.put<ApiResponse<{ id: number }>>(`/api/org-nodes/${nodeId}`, body, { params: { tenant_id: tenantId } }),
  )
}

export async function deleteOrgNode(tenantId: number, nodeId: number) {
  return unwrap(
    http.delete<ApiResponse<{ deleted: number }>>(`/api/org-nodes/${nodeId}`, { params: { tenant_id: tenantId } }),
  )
}

export async function createCompany(tenantId: number, body: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/companies', body, { params: { tenant_id: tenantId } }))
}

export async function updateCompany(tenantId: number, companyId: number, body: Record<string, unknown>) {
  return unwrap(
    http.put<ApiResponse<{ id: number }>>(`/api/companies/${companyId}`, body, { params: { tenant_id: tenantId } }),
  )
}

export async function createDepartment(tenantId: number, body: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/departments', body, { params: { tenant_id: tenantId } }))
}

export async function updateDepartment(tenantId: number, departmentId: number, body: Record<string, unknown>) {
  return unwrap(
    http.put<ApiResponse<{ id: number }>>(`/api/departments/${departmentId}`, body, {
      params: { tenant_id: tenantId },
    }),
  )
}

export async function createStore(tenantId: number, body: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/stores', body, { params: { tenant_id: tenantId } }))
}

export async function updateStore(tenantId: number, storeId: number, body: Record<string, unknown>) {
  return unwrap(
    http.put<ApiResponse<{ id: number }>>(`/api/stores/${storeId}`, body, { params: { tenant_id: tenantId } }),
  )
}

export async function deleteCompany(tenantId: number, companyId: number) {
  return unwrap(
    http.delete<ApiResponse<{ deleted: number }>>(`/api/companies/${companyId}`, { params: { tenant_id: tenantId } }),
  )
}

export async function deleteDepartment(tenantId: number, departmentId: number) {
  return unwrap(
    http.delete<ApiResponse<{ deleted: number }>>(`/api/departments/${departmentId}`, {
      params: { tenant_id: tenantId },
    }),
  )
}

export async function deleteStore(tenantId: number, storeId: number) {
  return unwrap(
    http.delete<ApiResponse<{ deleted: number }>>(`/api/stores/${storeId}`, { params: { tenant_id: tenantId } }),
  )
}
