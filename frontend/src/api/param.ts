import http, { unwrap } from './http'
import type { ApiResponse } from './types'
import type { Paginated } from './types'

export interface SysParamRow {
  id: number
  tenant_id?: number
  param_key: string
  default_value?: string | null
  param_value: string | null
  remark: string | null
  value_type?: string
  tenant_editable?: boolean
  is_platform_only?: boolean
  is_tenant_owned?: boolean
  is_override?: boolean
}

export async function fetchSysParams(skip = 0, limit = 20, opts?: { keyword?: string }) {
  const params: Record<string, string | number> = { skip, limit }
  const kw = (opts?.keyword ?? '').trim()
  if (kw) params.keyword = kw
  return unwrap(http.get<ApiResponse<Paginated<SysParamRow>>>('/api/sys-params', { params }))
}

/** 批量读取参数值（逗号分隔 key），用于前端默认值等 */
export async function fetchSysParamBatch(keys: string[]) {
  const keysParam = keys.filter(Boolean).join(',')
  return unwrap(
    http.get<ApiResponse<{ values: Record<string, string | null> }>>('/api/sys-params/batch', {
      params: keysParam ? { keys: keysParam } : {},
    }),
  )
}

export async function createSysParam(body: {
  param_key: string
  param_value?: string
  remark?: string
  value_type?: string
  tenant_editable?: boolean
  is_platform_only?: boolean
}) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/sys-params', body))
}

export async function updateSysParam(id: number, body: {
  param_value?: string | null
  remark?: string | null
  value_type?: string
  tenant_editable?: boolean
  is_platform_only?: boolean
}) {
  return unwrap(http.put<ApiResponse<SysParamRow>>(`/api/sys-params/${id}`, body))
}

export async function deleteSysParam(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/sys-params/${id}`))
}

export async function restoreSysParamDefault(id: number) {
  return unwrap(http.delete<ApiResponse<SysParamRow>>(`/api/sys-params/${id}/override`))
}
