import http, { unwrap } from './http'
import type { ApiResponse, Paginated } from './types'

export interface BusinessResourceRow {
  id: number
  tenant_id: number
  unit_type_code: string
  unit_type_name: string
  business_unit_code: string
  business_unit_name: string
  resource_name: string
  resource_code: string
  resource_category: string
  resource_type: string
  source_mode: string
  source_app_code: string | null
  source_table: string | null
  source_id: number | null
  platform_code: string | null
  external_id: string | null
  connection_instance_id: number | null
  parent_resource_id: number | null
  resource_attrs: string | null
  status: string
}

export interface BusinessUnitResourceRow {
  id: number
  tenant_id: number
  business_unit_id: number
  resource_id: number
  resource_category: string
  resource_type: string
  relation_type: string
  is_primary: boolean
  use_for_permission: boolean
  use_for_operation: boolean
  use_for_settlement: boolean
  start_date: string | null
  end_date: string | null
}

export interface BusinessResourceRelationRow {
  id: number
  tenant_id: number
  parent_resource_id: number
  child_resource_id: number
  parent_resource_category: string
  parent_resource_type: string
  child_resource_category: string
  child_resource_type: string
  relation_type: string
  start_date: string | null
  end_date: string | null
  status: string
}

export type BusinessResourceBody = {
  unit_type_code?: string
  business_unit_code?: string
  resource_name: string
  resource_code: string
  resource_category?: string
  resource_type?: string
  source_mode: string
  source_app_code?: string | null
  source_table?: string | null
  source_id?: number | null
  platform_code?: string | null
  external_id?: string | null
  connection_instance_id?: number | null
  parent_resource_id?: number | null
  resource_attrs?: Record<string, unknown> | null
  status: string
}

export interface BusinessUnitTreeNode {
  id: number
  code: string
  name: string
  sort_order: number
  children: BusinessUnitTreeNode[]
}

export interface BusinessResourceSummaryUnit {
  code: string
  name: string
  resource_count: number
}

export interface BusinessResourceSummaryType {
  code: string
  name: string
  resource_count: number
  business_units: BusinessResourceSummaryUnit[]
}

export interface BusinessResourceActorRow {
  id: number
  resource_id: number
  actor_type: 'org_node' | 'user'
  actor_id: number
  role_type: string
  include_children: boolean
  status: string
}

export interface BusinessResourceFieldConfig {
  field_key: string
  field_label: string
  field_type: string
  dict_code?: string | null
  required: boolean
  placeholder?: string | null
  help_text?: string | null
  show_in_list: boolean
  show_in_detail: boolean
  show_in_import: boolean
  sort_order: number
}

export async function fetchBusinessUnitTreeDict() {
  return unwrap(http.get<ApiResponse<{ dict_code: string; dict_name: string; children: BusinessUnitTreeNode[] }>>('/api/base/business-resources/dictionary/tree'))
}

export async function fetchBusinessResourceSummary() {
  return unwrap(http.get<ApiResponse<{ items: BusinessResourceSummaryType[] }>>('/api/base/business-resources/summary'))
}

export async function fetchBusinessResources(params: {
  skip?: number
  limit?: number
  keyword?: string
  unit_type_code?: string
  business_unit_code?: string
  resource_category?: string
  resource_type?: string
  source_mode?: string
  status?: string
}) {
  return unwrap(http.get<ApiResponse<Paginated<BusinessResourceRow>>>('/api/base/business-resources', { params }))
}

export async function createBusinessResource(body: BusinessResourceBody) {
  return unwrap(http.post<ApiResponse<BusinessResourceRow>>('/api/base/business-resources', body))
}

export async function updateBusinessResource(id: number, body: BusinessResourceBody) {
  return unwrap(http.put<ApiResponse<BusinessResourceRow>>(`/api/base/business-resources/${id}`, body))
}

export async function deleteBusinessResource(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/base/business-resources/${id}`))
}

export async function fetchBusinessResourceActors(resourceId: number) {
  return unwrap(http.get<ApiResponse<BusinessResourceActorRow[]>>(`/api/base/business-resources/${resourceId}/actors`))
}

export async function saveBusinessResourceActors(resourceId: number, actors: Array<{
  actor_type: 'org_node' | 'user'
  actor_id: number
  role_type: string
  include_children?: boolean
  status?: string
}>) {
  return unwrap(http.put<ApiResponse<BusinessResourceActorRow[]>>(`/api/base/business-resources/${resourceId}/actors`, { actors }))
}

export async function fetchBusinessResourceFieldConfigs(params: { unit_type_code: string; business_unit_code: string }) {
  return unwrap(http.get<ApiResponse<{ unit_type_code: string; business_unit_code: string; fields: BusinessResourceFieldConfig[] }>>('/api/base/business-resources/field-configs', { params }))
}

export async function fetchBusinessUnitResources(businessUnitId: number) {
  return unwrap(http.get<ApiResponse<BusinessUnitResourceRow[]>>(`/api/business-units/${businessUnitId}/resources`))
}

export async function bindBusinessUnitResource(
  businessUnitId: number,
  body: {
    resource_id: number
    relation_type: string
    is_primary?: boolean
    use_for_permission?: boolean
    use_for_operation?: boolean
    use_for_settlement?: boolean
    start_date?: string | null
    end_date?: string | null
  },
) {
  return unwrap(http.post<ApiResponse<BusinessUnitResourceRow>>(`/api/business-units/${businessUnitId}/resources`, body))
}

export async function deleteBusinessUnitResource(businessUnitId: number, relationId: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/business-units/${businessUnitId}/resources/${relationId}`))
}

export async function fetchBusinessResourceRelations(resourceId: number) {
  return unwrap(http.get<ApiResponse<BusinessResourceRelationRow[]>>(`/api/base/business-resources/${resourceId}/relations`))
}

export async function createBusinessResourceRelation(
  resourceId: number,
  body: { child_resource_id: number; relation_type: string; start_date?: string | null; end_date?: string | null; status?: string },
) {
  return unwrap(http.put<ApiResponse<BusinessResourceRelationRow[]>>(`/api/base/business-resources/${resourceId}/relations`, { relations: [{ target_resource_id: body.child_resource_id, relation_type: body.relation_type, start_date: body.start_date, end_date: body.end_date, status: body.status }] }))
}

export async function deleteBusinessResourceRelation(resourceId: number, relationId: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/business-resources/${resourceId}/relations/${relationId}`))
}
