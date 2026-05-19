import http, { unwrap } from './http'
import type { ApiResponse, Paginated } from './types'

/** 旧版业务单元接口保留兼容；新版页面使用下方 /api/base/business-units 接口。 */
export interface BusinessUnitRow {
  id: number
  tenant_id: number
  name: string
  code: string
  bu_type: string | null
  unit_scenario: string | null
  unit_form: string | null
  parent_id: number | null
  owner_user_id: number | null
  owner_org_id: number | null
  status: number
  billing_enabled: boolean
  statistic_enabled: boolean
  operation_enabled: boolean
  settlement_enabled: boolean
  data_scope_enabled: boolean
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
  bu_type?: string
  unit_scenario?: string | null
  unit_form: string
  parent_id?: number | null
  owner_user_id?: number | null
  operation_enabled?: boolean
  settlement_enabled?: boolean
  data_scope_enabled?: boolean
  org_node_ids?: number[]
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
    unit_scenario: string | null
    unit_form: string | null
    parent_id: number | null
    owner_user_id: number | null
    operation_enabled: boolean
    settlement_enabled: boolean
    data_scope_enabled: boolean
    status: number
    remark: string | null
    org_node_ids?: number[]
  }>,
) {
  return unwrap(http.put<ApiResponse<{ id: number }>>(`/api/business-units/${id}`, body))
}

export async function deleteBusinessUnit(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/business-units/${id}`))
}

export interface BusinessUnitDictNode {
  id: number
  code: string
  name: string
  sort_order: number
  children: BusinessUnitDictNode[]
}

export interface BusinessUnitSummaryGroup {
  code: string
  name: string
  unit_count: number
}

export interface BusinessUnitSummaryType {
  code: string
  name: string
  unit_count: number
  groups: BusinessUnitSummaryGroup[]
}

export interface BusinessUnitV3Row {
  id: number
  tenant_id: number
  name: string
  code: string
  unit_type_code: string | null
  unit_type_name: string | null
  unit_group_code: string | null
  unit_group_name: string | null
  parent_id: number | null
  attr_template_id: number | null
  attrs: string | null
  status: number
  remark: string | null
}

export interface BusinessUnitActorRow {
  id: number
  business_unit_id: number
  actor_type: 'org' | 'user'
  actor_id: number
  role_type: string
  include_children: boolean
  status: string
}

export interface BusinessUnitRelationRow {
  id: number
  source_unit_id: number
  target_unit_id: number
  relation_type_code: string
  relation_type_name: string
  status: string
  remark: string | null
}

export interface BusinessUnitAttrTemplate {
  id: number
  template_name: string
  unit_type_code: string
  unit_type_name: string
  unit_group_code: string | null
  unit_group_name: string | null
  sort_order: number
  status: string
  remark: string | null
}

export interface BusinessUnitAttrField {
  id: number
  template_id: number
  field_key: string
  field_label: string
  field_type: 'text' | 'number' | 'date' | 'select' | 'textarea' | 'switch'
  required: boolean
  default_value: string | null
  placeholder: string | null
  options_json: string | null
  sort_order: number
  status: string
}

export type BusinessUnitAttrFieldBody = {
  field_key: string
  field_label: string
  field_type: BusinessUnitAttrField['field_type']
  required?: boolean
  default_value?: string | null
  placeholder?: string | null
  options_json?: unknown
  sort_order?: number
  status?: string
}

export type BusinessUnitAttrTemplateBody = {
  template_name: string
  unit_type_code: string
  unit_group_code?: string | null
  sort_order?: number
  status?: string
  remark?: string | null
  fields: BusinessUnitAttrFieldBody[]
}

export type BusinessUnitV3Body = {
  unit_type_code: string
  unit_group_code: string
  name: string
  code: string
  parent_id?: number | null
  attr_template_id?: number | null
  attrs?: Record<string, unknown> | null
  status: number
  remark?: string | null
}

export async function fetchBusinessUnitDictionaryTree() {
  return unwrap(http.get<ApiResponse<{ dict_code: string; dict_name: string; children: BusinessUnitDictNode[] }>>('/api/base/business-units/dictionary/tree'))
}

export async function fetchBusinessUnitSummary() {
  return unwrap(http.get<ApiResponse<{ items: BusinessUnitSummaryType[] }>>('/api/base/business-units/summary'))
}

export async function fetchBusinessUnitV3Page(params: {
  skip?: number
  limit?: number
  keyword?: string
  unit_type_code?: string
  unit_group_code?: string
  status?: number
}) {
  return unwrap(http.get<ApiResponse<Paginated<BusinessUnitV3Row>>>('/api/base/business-units', { params }))
}

export async function createBusinessUnitV3(body: BusinessUnitV3Body) {
  return unwrap(http.post<ApiResponse<BusinessUnitV3Row>>('/api/base/business-units', body))
}

export async function updateBusinessUnitV3(id: number, body: BusinessUnitV3Body) {
  return unwrap(http.put<ApiResponse<BusinessUnitV3Row>>(`/api/base/business-units/${id}`, body))
}

export async function deleteBusinessUnitV3(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/base/business-units/${id}`))
}

export async function fetchBusinessUnitActors(unitId: number) {
  return unwrap(http.get<ApiResponse<BusinessUnitActorRow[]>>(`/api/base/business-units/${unitId}/actors`))
}

export async function saveBusinessUnitActors(unitId: number, actors: Array<{
  actor_type: 'org' | 'user'
  actor_id: number
  role_type: string
  include_children?: boolean
  status?: string
}>) {
  return unwrap(http.put<ApiResponse<BusinessUnitActorRow[]>>(`/api/base/business-units/${unitId}/actors`, { actors }))
}

export async function fetchBusinessUnitRelations(unitId: number) {
  return unwrap(http.get<ApiResponse<BusinessUnitRelationRow[]>>(`/api/base/business-units/${unitId}/relations`))
}

export async function saveBusinessUnitRelations(unitId: number, relations: Array<{
  target_unit_id: number
  relation_type_code: string
  relation_type_name?: string
  status?: string
  remark?: string | null
}>) {
  return unwrap(http.put<ApiResponse<BusinessUnitRelationRow[]>>(`/api/base/business-units/${unitId}/relations`, { relations }))
}

export async function matchBusinessUnitAttrTemplate(params: { unit_type_code: string; unit_group_code?: string }) {
  return unwrap(http.get<ApiResponse<{ template: BusinessUnitAttrTemplate | null; fields: BusinessUnitAttrField[] }>>('/api/base/business-unit-attr-templates/match', { params }))
}

export async function fetchBusinessUnitAttrTemplates(params?: { unit_type_code?: string; unit_group_code?: string; status?: string }) {
  const data = await unwrap(http.get<ApiResponse<{ items: BusinessUnitAttrTemplate[] }>>('/api/base/business-unit-attr-templates', { params }))
  return data.items
}

export async function fetchBusinessUnitAttrTemplate(id: number) {
  return unwrap(http.get<ApiResponse<{ template: BusinessUnitAttrTemplate; fields: BusinessUnitAttrField[] }>>(`/api/base/business-unit-attr-templates/${id}`))
}

export async function createBusinessUnitAttrTemplate(body: BusinessUnitAttrTemplateBody) {
  return unwrap(http.post<ApiResponse<{ template: BusinessUnitAttrTemplate; fields: BusinessUnitAttrField[] }>>('/api/base/business-unit-attr-templates', body))
}

export async function updateBusinessUnitAttrTemplate(id: number, body: BusinessUnitAttrTemplateBody) {
  return unwrap(http.put<ApiResponse<{ template: BusinessUnitAttrTemplate; fields: BusinessUnitAttrField[] }>>(`/api/base/business-unit-attr-templates/${id}`, body))
}

export async function deleteBusinessUnitAttrTemplate(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/base/business-unit-attr-templates/${id}`))
}
