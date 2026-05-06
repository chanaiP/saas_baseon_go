import http, { unwrap } from './http'
import type { ApiResponse } from './types'
import type { Paginated } from './types'

export interface DictTypeRow {
  id: number
  code: string
  name: string
  remark: string | null
  scope?: string
  tenant_editable?: boolean
  is_platform_only?: boolean
}

export interface DictItemRow {
  id: number
  dict_type_id: number
  default_label?: string
  default_value?: string
  default_sort_order?: number
  default_enabled?: boolean
  label: string
  value: string
  sort_order: number
  enabled?: boolean
  is_override?: boolean
}

export async function fetchDictTypes(
  skip = 0,
  limit = 20,
  opts?: { keyword?: string; platform_only?: boolean },
) {
  const params: Record<string, string | number | boolean> = { skip, limit }
  const kw = (opts?.keyword ?? '').trim()
  if (kw) params.keyword = kw
  if (opts?.platform_only === true) params.platform_only = true
  if (opts?.platform_only === false) params.platform_only = false
  return unwrap(http.get<ApiResponse<Paginated<DictTypeRow>>>('/api/dict-types', { params }))
}

/** 按类型编码拉取字典项（供表单下拉等，与数据字典维护中的 code 一致） */
export async function fetchDictItemsByCode(code: string) {
  return unwrap(
    http.get<ApiResponse<{ code: string; items: DictItemRow[] }>>(`/api/dict-types/by-code/${encodeURIComponent(code)}/items`),
  )
}

export async function createDictType(body: {
  code: string
  name: string
  remark?: string
  scope?: string
  tenant_editable?: boolean
  is_platform_only?: boolean
}) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/dict-types', body))
}

export async function updateDictType(id: number, body: {
  name?: string
  remark?: string | null
  scope?: string
  tenant_editable?: boolean
  is_platform_only?: boolean
}) {
  return unwrap(http.put<ApiResponse<DictTypeRow>>(`/api/dict-types/${id}`, body))
}

export async function deleteDictType(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/dict-types/${id}`))
}

export async function fetchDictItems(dictTypeId: number, skip = 0, limit = 20) {
  return unwrap(
    http.get<ApiResponse<Paginated<DictItemRow>>>('/api/dict-items', {
      params: { dict_type_id: dictTypeId, skip, limit },
    }),
  )
}

export async function createDictItem(body: {
  dict_type_id: number
  label: string
  value: string
  sort_order?: number
  enabled?: boolean
}) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/dict-items', body))
}

export async function updateDictItem(
  id: number,
  body: { label?: string; value?: string; sort_order?: number; enabled?: boolean },
) {
  return unwrap(http.put<ApiResponse<DictItemRow>>(`/api/dict-items/${id}`, body))
}

export async function deleteDictItem(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/dict-items/${id}`))
}

export async function restoreDictItemDefault(id: number) {
  return unwrap(http.delete<ApiResponse<DictItemRow>>(`/api/dict-items/${id}/override`))
}
