import http, { unwrap } from './http'
import type { ApiResponse, Paginated } from './types'

/** 一级业务单元主数据（无父子嵌套；与组织多对多通过 org-mappings） */
export interface BusinessUnitRow {
  id: number
  tenant_id: number
  name: string
  code: string
  bu_type: string | null
  status: number
  billing_enabled: boolean
  statistic_enabled: boolean
  remark: string | null
}

/** 与 /tree 同源扁平列表，保留类型别名便于调用方阅读 */
export type BusinessUnitFlatNode = BusinessUnitRow

export interface BusinessUnitOrgMapRow {
  id: number
  tenant_id: number
  business_unit_id: number
  org_id: number
  org_type: string
  scope_type: string
  priority: number
  status: number
}

export async function fetchBusinessUnitPage(params: {
  skip?: number
  limit?: number
  keyword?: string
  status?: number
}) {
  return unwrap(http.get<ApiResponse<Paginated<BusinessUnitRow>>>('/api/business-units', { params }))
}

/** 扁平列表；路径为旧版「树」接口的兼容别名 */
export async function fetchBusinessUnitTree() {
  return unwrap(http.get<ApiResponse<BusinessUnitFlatNode[]>>('/api/business-units/tree'))
}

export async function fetchBusinessUnitOrgMappings(buId: number) {
  return unwrap(http.get<ApiResponse<BusinessUnitOrgMapRow[]>>(`/api/business-units/${buId}/org-mappings`))
}

export async function addBusinessUnitOrgMapping(buId: number, body: { org_id: number }) {
  return unwrap(http.post<ApiResponse<{ id: number }>>(`/api/business-units/${buId}/org-mappings`, body))
}

export async function deleteBusinessUnitOrgMapping(mapId: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/business-units/org-mappings/${mapId}`))
}

export async function createBusinessUnit(body: {
  name: string
  code: string
  bu_type: string
  org_node_ids: number[]
  status: number
  remark?: string | null
}) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/business-units', body))
}

export async function updateBusinessUnit(
  id: number,
  body: Partial<{
    name: string
    code: string
    bu_type: string | null
    status: number
    remark: string | null
    org_node_ids: number[]
  }>,
) {
  return unwrap(http.put<ApiResponse<{ id: number }>>(`/api/business-units/${id}`, body))
}

export async function deleteBusinessUnit(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/business-units/${id}`))
}
